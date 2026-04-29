package assinador

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hubsaude/runner/assinatura/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Garantir reusa um servidor existente quando state.json aponta para um PID
// vivo respondendo no /health.
func TestGarantir_ReusaInstanciaExistente(t *testing.T) {
	dir := redirecionarHubsaude(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"UP"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	porta := portaDe(t, srv.URL)

	st := &state.State{Assinador: &state.ProcessInfo{
		PID:        os.Getpid(),
		Porta:      porta,
		IniciadoEm: time.Now().UTC(),
		JavaPath:   "java",
	}}
	require.NoError(t, st.Save())

	var log bytes.Buffer
	s, err := Garantir(context.Background(), OpcoesStartup{
		JarPath:        "fake.jar",
		PortaPreferida: porta,
		LogProgresso:   &log,
		TimeoutPronto:  500 * time.Millisecond,
	})
	require.NoError(t, err)
	assert.True(t, s.Reusado)
	assert.Equal(t, porta, s.Porta)
	assert.Contains(t, log.String(), "em execucao")
	_ = dir
}

func TestGarantir_DescartaPidObsoleto(t *testing.T) {
	redirecionarHubsaude(t)
	st := &state.State{Assinador: &state.ProcessInfo{
		PID:        99999999,
		Porta:      8088,
		IniciadoEm: time.Now().UTC(),
	}}
	require.NoError(t, st.Save())

	_, err := Garantir(context.Background(), OpcoesStartup{
		JarPath:        "/nao-existe/fake.jar",
		PortaPreferida: 8088,
		JavaPath:       "/binario/inexistente",
		TimeoutPronto:  100 * time.Millisecond,
	})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "reusar")
}

func TestGarantir_JarPathObrigatorio(t *testing.T) {
	_, err := Garantir(context.Background(), OpcoesStartup{})
	assert.ErrorContains(t, err, "JarPath obrigatorio")
}

// redirecionarHubsaude aponta HOME (e APPDATA no Windows) para um diretorio
// temporario, isolando state.json/config.json dos testes.
func redirecionarHubsaude(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("APPDATA", dir)
	return dir
}

func portaDe(t *testing.T, url string) int {
	t.Helper()
	u := strings.TrimPrefix(url, "http://")
	idx := strings.LastIndex(u, ":")
	require.True(t, idx > 0)
	p, err := strconv.Atoi(u[idx+1:])
	require.NoError(t, err)
	return p
}
