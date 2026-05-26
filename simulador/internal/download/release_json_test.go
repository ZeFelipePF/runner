package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuscarReleaseJson_HTTPSucesso(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"jar": {"url": "https://example.com/simulador.jar", "version": "1.5.0", "sha256": "abc123"},
			"jre": {"linux_x64": "https://example.com/jre.tar.gz"}
		}`))
	}))
	defer srv.Close()

	rj, err := BuscarReleaseJson(context.Background(), http.DefaultClient, srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "1.5.0", rj.Jar.Version)
	assert.Equal(t, "https://example.com/simulador.jar", rj.Jar.URL)
	assert.Equal(t, "abc123", rj.Jar.SHA256)
	assert.Equal(t, "https://example.com/jre.tar.gz", rj.JRE["linux_x64"])
}

func TestBuscarReleaseJson_FileURL(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "release.json")
	require.NoError(t, os.WriteFile(path,
		[]byte(`{"jar":{"url":"file:///tmp/jar","version":"x"}}`), 0o644))
	rj, err := BuscarReleaseJson(context.Background(), http.DefaultClient, "file://"+path)
	require.NoError(t, err)
	assert.Equal(t, "x", rj.Jar.Version)
}

func TestBuscarReleaseJson_JSONInvalido(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()
	_, err := BuscarReleaseJson(context.Background(), http.DefaultClient, srv.URL)
	require.Error(t, err)
}

func TestBuscarReleaseJson_SemJarURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jar":{"version":"1.0"}}`))
	}))
	defer srv.Close()
	_, err := BuscarReleaseJson(context.Background(), http.DefaultClient, srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jar.url")
}

func TestURLReleaseJsonOuPadrao_RespeitaEnv(t *testing.T) {
	t.Setenv("HUBSAUDE_RELEASE_JSON_URL", "https://meu/release.json")
	assert.Equal(t, "https://meu/release.json", URLReleaseJsonOuPadrao())
}

func TestGarantir_ComReleaseJsonComoPrimario(t *testing.T) {
	jar := []byte("conteudo do simulador.jar via release.json")
	soma := sha256.Sum256(jar)

	dlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(jar)
	}))
	defer dlSrv.Close()

	rjSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jar":{"url":"` + dlSrv.URL + `/simulador.jar","version":"v3.0.0"}}`))
	}))
	defer rjSrv.Close()

	tmp := t.TempDir()
	r, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir:    tmp,
		ReleaseJsonURL: rjSrv.URL,
	})
	require.NoError(t, err)
	assert.Equal(t, "v3.0.0", r.Metadados.Versao)
	assert.Equal(t, hex.EncodeToString(soma[:]), r.Metadados.SHA256)
}

func TestGarantir_ReleaseJsonFalhaCaiNaGitHubAPI(t *testing.T) {
	jar := []byte("fallback via github")
	soma := sha256.Sum256(jar)
	ghSrv := mockGithubServer(t, jar, "v2.0.0")
	defer ghSrv.Close()

	// release.json deliberadamente 404
	rjSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer rjSrv.Close()

	t.Setenv("HUBSAUDE_SIMULADOR_REPO", ghSrv.URL+"/releases")
	tmp := t.TempDir()
	r, err := Garantir(context.Background(), Opcoes{
		HubsaudeDir:    tmp,
		ReleaseJsonURL: rjSrv.URL,
	})
	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", r.Metadados.Versao)
	assert.Equal(t, hex.EncodeToString(soma[:]), r.Metadados.SHA256)
}
