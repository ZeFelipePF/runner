package jdk

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMajor(t *testing.T) {
	casos := map[string]int{
		"21.0.3":     21,
		"17.0.10":    17,
		"1.8.0_421":  8,
		"22-ea":      22,
	}
	for s, esperado := range casos {
		got, err := parseMajor(s)
		require.NoError(t, err, s)
		assert.Equal(t, esperado, got, s)
	}
}

func TestDetectar_UsaPATHQuandoDisponivel(t *testing.T) {
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java nao disponivel no PATH")
	}
	// Garante que JAVA_HOME nao influencia.
	t.Setenv("JAVA_HOME", "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	j, err := Detectar(Opcoes{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, j.VersaoMajor, VersaoMinima)
	assert.Contains(t, []string{"PATH", "JAVA_HOME"}, j.Origem)
}

func TestDetectar_NaoEncontradoQuandoTudoAusente(t *testing.T) {
	t.Setenv("JAVA_HOME", "")
	t.Setenv("PATH", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("APPDATA", tmp)

	_, err := Detectar(Opcoes{HubsaudeDir: tmp})
	assert.ErrorIs(t, err, ErrNaoEncontrado)
}

func TestDetectar_DescobreJDKLocal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("teste cria binario shell-script tipo unix")
	}
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java nao disponivel no PATH para reuso")
	}

	javaReal, _ := exec.LookPath("java")
	tmp := t.TempDir()
	releaseDir := filepath.Join(tmp, "jdk", "temurin-21-test", "bin")
	require.NoError(t, os.MkdirAll(releaseDir, 0o755))
	link := filepath.Join(releaseDir, "java")
	require.NoError(t, os.Symlink(javaReal, link))

	t.Setenv("JAVA_HOME", "")
	t.Setenv("PATH", "")

	j, err := Detectar(Opcoes{HubsaudeDir: tmp})
	require.NoError(t, err)
	assert.Equal(t, "hubsaude", j.Origem)
	assert.GreaterOrEqual(t, j.VersaoMajor, VersaoMinima)
}

// Teste de Garantir com Adoptium falso usando httptest.
// Constroi um JDK fake: um tar.gz contendo bin/java que e um script shell que
// reporta versao 21.0.0.
func TestGarantir_BaixaQuandoAusente(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("teste de extracao tar.gz; pular Windows")
	}

	jdkDir := t.TempDir()
	fakeJar := criarJDKFake(t)
	checksum := sha256Hex(t, fakeJar)
	dadosJar, err := os.ReadFile(fakeJar)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/assets/latest/21/hotspot", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{{
			"release_name": "jdk-21.0.0-fake",
			"binary": map[string]any{
				"package": map[string]any{
					"link":     "http://" + r.Host + "/download",
					"checksum": checksum,
					"name":     "OpenJDK21U-jdk_x64_fake.tar.gz",
				},
			},
			"version": map[string]any{"semver": "21.0.0", "major": 21},
		}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Write(dadosJar)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("HUBSAUDE_ADOPTIUM_URL", srv.URL)
	t.Setenv("JAVA_HOME", "")
	t.Setenv("PATH", "")

	j, err := Garantir(context.Background(), Opcoes{HubsaudeDir: jdkDir})
	require.NoError(t, err)
	assert.Equal(t, "adoptium", j.Origem)
	assert.Equal(t, 21, j.VersaoMajor)
	assert.FileExists(t, j.JavaBin)
	assert.FileExists(t, filepath.Join(jdkDir, "jdk", "jdk-21.0.0-fake", "hubsaude-meta.json"))
}

// criarJDKFake cria um tar.gz com layout `jdk-fake/bin/java` (script shell).
func criarJDKFake(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "jdk.tar.gz")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	script := "#!/bin/sh\necho 'openjdk version \"21.0.0\" 2024-01-01' 1>&2\n"
	hdr := &tar.Header{Name: "jdk-fake/bin/java", Mode: 0o755, Size: int64(len(script)), Typeflag: tar.TypeReg}
	require.NoError(t, tw.WriteHeader(hdr))
	_, err = io.Copy(tw, strings.NewReader(script))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return path
}

func sha256Hex(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	require.NoError(t, err)
	return hex.EncodeToString(h.Sum(nil))
}

func TestGarantir_FalhaComChecksumDivergente(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	dir := t.TempDir()
	dados := []byte("conteudo-fake")
	mux := http.NewServeMux()
	mux.HandleFunc("/assets/latest/21/hotspot", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{{
			"release_name": "jdk-fake",
			"binary": map[string]any{
				"package": map[string]any{
					"link":     "http://" + r.Host + "/dl",
					"checksum": "deadbeef",
				},
			},
			"version": map[string]any{"semver": "21.0.0", "major": 21},
		}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) { w.Write(dados) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("HUBSAUDE_ADOPTIUM_URL", srv.URL)
	t.Setenv("JAVA_HOME", "")
	t.Setenv("PATH", "")
	_, err := Garantir(context.Background(), Opcoes{HubsaudeDir: dir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum")
}

func TestPlataformaAdoptium(t *testing.T) {
	osN, arch, ext, err := plataformaAdoptium()
	require.NoError(t, err)
	assert.NotEmpty(t, osN)
	assert.NotEmpty(t, arch)
	assert.True(t, ext == "tar.gz" || ext == "zip")
}

func TestDescascarPrefixo(t *testing.T) {
	assert.Equal(t, "bin/java", descascarPrefixo("jdk-21/bin/java"))
	assert.Equal(t, "bin/java", descascarPrefixo("./jdk-21/bin/java"))
	assert.Equal(t, "", descascarPrefixo("README"))
}

// silencer para evitar warnings de import nao usado em algumas plataformas
var _ = fmt.Sprintf
