package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockGithubServer(t *testing.T, jarBytes []byte, tag string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"tag_name": tag,
			"assets": []map[string]string{{
				"name":                 "simulador.jar",
				"browser_download_url": "http://" + r.Host + "/dl/simulador.jar",
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/releases/tags/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"tag_name": tag,
			"assets": []map[string]string{{
				"name":                 "simulador.jar",
				"browser_download_url": "http://" + r.Host + "/dl/simulador.jar",
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/dl/simulador.jar", func(w http.ResponseWriter, r *http.Request) {
		w.Write(jarBytes)
	})
	return httptest.NewServer(mux)
}

func TestGarantir_BaixaELatentaCache(t *testing.T) {
	jar := []byte("conteudo do simulador.jar v1")
	soma := sha256.Sum256(jar)
	somaHex := hex.EncodeToString(soma[:])
	srv := mockGithubServer(t, jar, "v1.0.0")
	defer srv.Close()

	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")
	dir := t.TempDir()

	r1, err := Garantir(context.Background(), Opcoes{HubsaudeDir: dir})
	require.NoError(t, err)
	assert.False(t, r1.UsouCache)
	assert.Equal(t, "v1.0.0", r1.Metadados.Versao)
	assert.Equal(t, somaHex, r1.Metadados.SHA256)
	assert.FileExists(t, r1.JarPath)
	assert.Equal(t, int64(len(jar)), r1.Metadados.TamanhoB)

	// Segunda chamada: reusa cache.
	r2, err := Garantir(context.Background(), Opcoes{HubsaudeDir: dir})
	require.NoError(t, err)
	assert.True(t, r2.UsouCache)
}

func TestGarantir_VersaoEspecifica(t *testing.T) {
	jar := []byte("conteudo v2")
	srv := mockGithubServer(t, jar, "v2.0.0")
	defer srv.Close()
	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")
	dir := t.TempDir()

	r, err := Garantir(context.Background(), Opcoes{HubsaudeDir: dir, Versao: "v2.0.0"})
	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", r.Metadados.Versao)
}

func TestGarantir_SHA256Esperado_OK(t *testing.T) {
	jar := []byte("payload xyz")
	soma := sha256.Sum256(jar)
	somaHex := hex.EncodeToString(soma[:])
	srv := mockGithubServer(t, jar, "v3")
	defer srv.Close()
	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")

	r, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir:    t.TempDir(),
		SHA256Esperado: somaHex,
	})
	require.NoError(t, err)
	assert.Equal(t, somaHex, r.Metadados.SHA256)
}

func TestGarantir_SHA256Divergente_Falha(t *testing.T) {
	jar := []byte("conteudo")
	srv := mockGithubServer(t, jar, "v1")
	defer srv.Close()
	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")

	dir := t.TempDir()
	_, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir:    dir,
		SHA256Esperado: "deadbeef",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SHA-256")
	// jar nao deve permanecer.
	_, statErr := os.Stat(filepath.Join(dir, "simulador", NomeArquivo))
	assert.Error(t, statErr)
}

func TestGarantir_SourceCustom_HTTP(t *testing.T) {
	jar := []byte("dev build")
	soma := sha256.Sum256(jar)
	somaHex := hex.EncodeToString(soma[:])
	mux := http.NewServeMux()
	mux.HandleFunc("/simulador-dev.jar", func(w http.ResponseWriter, r *http.Request) {
		w.Write(jar)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	r, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir: t.TempDir(),
		Source:      srv.URL + "/simulador-dev.jar",
	})
	require.NoError(t, err)
	assert.Equal(t, "custom", r.Metadados.Versao)
	assert.Equal(t, somaHex, r.Metadados.SHA256)
}

func TestGarantir_SourceCustom_FileURL(t *testing.T) {
	jar := []byte("local build")
	soma := sha256.Sum256(jar)
	somaHex := hex.EncodeToString(soma[:])
	src := filepath.Join(t.TempDir(), "src.jar")
	require.NoError(t, os.WriteFile(src, jar, 0o644))

	r, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir: t.TempDir(),
		Source:      "file://" + src,
	})
	require.NoError(t, err)
	assert.Equal(t, somaHex, r.Metadados.SHA256)
}

func TestGarantir_ForcarRedownload_IgnoraCache(t *testing.T) {
	jarV1 := []byte("v1-content")
	tag := "v1"
	var contador int
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"tag_name": tag,
			"assets": []map[string]string{{
				"name":                 "simulador.jar",
				"browser_download_url": fmt.Sprintf("http://%s/dl", r.Host),
			}},
		})
	})
	mux.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		contador++
		w.Write(jarV1)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")

	dir := t.TempDir()
	_, err := Garantir(context.Background(), Opcoes{HubsaudeDir: dir})
	require.NoError(t, err)
	_, err = Garantir(context.Background(), Opcoes{HubsaudeDir: dir, ForcarRedownload: true})
	require.NoError(t, err)
	assert.Equal(t, 2, contador, "ForcarRedownload deve forcar segundo download")
}

func TestGarantir_GitHubReleases_TagInexistente(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/tags/v999", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")

	_, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir: t.TempDir(),
		Versao:      "v999",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "v999")
}

func TestGarantir_NenhumAsset(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"tag_name": "v1", "assets": []any{}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("HUBSAUDE_SIMULADOR_REPO", srv.URL+"/releases")
	_, err := Garantir(context.Background(), Opcoes{HubsaudeDir: t.TempDir()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "asset")
}

func TestSelecionarAsset_FiltrarJar(t *testing.T) {
	rel := &releaseGitHub{TagName: "v1"}
	rel.Assets = append(rel.Assets, struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	}{Name: "checksums.txt", URL: "u1"})
	rel.Assets = append(rel.Assets, struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	}{Name: "simulador.jar", URL: "u2"})
	a, err := selecionarAsset(rel)
	require.NoError(t, err)
	assert.Equal(t, "u2", a.URL)
}
