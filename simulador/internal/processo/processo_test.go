package processo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hubsaude/runner/simulador/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func redirHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("APPDATA", dir)
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

func TestIniciar_ReusaInstanciaExistente(t *testing.T) {
	redirHome(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/actuator/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"UP"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	p := portaDe(t, srv.URL)

	st := &state.State{Simulador: &state.ProcessInfo{
		PID:        os.Getpid(),
		Porta:      p,
		IniciadoEm: time.Now().UTC(),
		JavaPath:   "java",
	}}
	require.NoError(t, st.Save())

	s, err := Iniciar(context.Background(), OpcoesIniciar{
		JarPath:       "fake.jar",
		PortaPreferida: p,
		TimeoutPronto:  500 * time.Millisecond,
	})
	require.NoError(t, err)
	assert.True(t, s.Reusado)
	assert.Equal(t, p, s.Porta)
}

func TestIniciar_JarObrigatorio(t *testing.T) {
	redirHome(t)
	_, err := Iniciar(context.Background(), OpcoesIniciar{})
	assert.ErrorContains(t, err, "JarPath obrigatorio")
}

func TestIniciar_DescartaPidObsoleto(t *testing.T) {
	redirHome(t)
	st := &state.State{Simulador: &state.ProcessInfo{
		PID:        99999999,
		Porta:      state.PortaPadraoSimulador,
		IniciadoEm: time.Now().UTC(),
	}}
	require.NoError(t, st.Save())

	_, err := Iniciar(context.Background(), OpcoesIniciar{
		JarPath:        "/nao-existe/fake.jar",
		PortaPreferida: state.PortaPadraoSimulador,
		JavaPath:       "/binario/inexistente",
		TimeoutPronto:  100 * time.Millisecond,
	})
	require.Error(t, err)
}

func TestConsultar_NaoRegistrado(t *testing.T) {
	redirHome(t)
	s, err := Consultar(context.Background())
	require.NoError(t, err)
	assert.False(t, s.Registrado)
}

func TestConsultar_PIDVivoEHealthOK(t *testing.T) {
	redirHome(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/actuator/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	p := portaDe(t, srv.URL)
	st := &state.State{Simulador: &state.ProcessInfo{
		PID:        os.Getpid(),
		Porta:      p,
		IniciadoEm: time.Now().UTC().Add(-time.Minute),
		Versao:     "v1.0.0",
	}}
	require.NoError(t, st.Save())
	s, err := Consultar(context.Background())
	require.NoError(t, err)
	assert.True(t, s.Registrado)
	assert.True(t, s.PIDVivo)
	assert.True(t, s.HealthOK)
	assert.Equal(t, "v1.0.0", s.Versao)
	assert.Greater(t, s.Uptime, 30*time.Second)
}

func TestConsultar_ApiInfoPopulado(t *testing.T) {
	redirHome(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/actuator/health":
			w.WriteHeader(http.StatusOK)
		case "/api/info":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"app":"hubsaude-simulador","versao":"1.2.3","env":"dev"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	p := portaDe(t, srv.URL)
	st := &state.State{Simulador: &state.ProcessInfo{
		PID:        os.Getpid(),
		Porta:      p,
		IniciadoEm: time.Now().UTC().Add(-time.Minute),
	}}
	require.NoError(t, st.Save())

	s, err := Consultar(context.Background())
	require.NoError(t, err)
	require.NotNil(t, s.Info)
	assert.Equal(t, "hubsaude-simulador", s.Info["app"])
	assert.Equal(t, "1.2.3", s.Info["versao"])
}

func TestConsultar_ApiInfoIndisponivelNaoFalha(t *testing.T) {
	redirHome(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/actuator/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// /api/info deliberadamente 404
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	p := portaDe(t, srv.URL)
	st := &state.State{Simulador: &state.ProcessInfo{
		PID:        os.Getpid(),
		Porta:      p,
		IniciadoEm: time.Now().UTC().Add(-time.Minute),
	}}
	require.NoError(t, st.Save())

	s, err := Consultar(context.Background())
	require.NoError(t, err)
	assert.True(t, s.PIDVivo)
	assert.True(t, s.HealthOK)
	assert.Nil(t, s.Info)
}

func TestParar_NaoEstavaRodando(t *testing.T) {
	redirHome(t)
	r, err := Parar(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "nao_estava_rodando", r.Metodo)
}

func TestParar_StalePid(t *testing.T) {
	redirHome(t)
	st := &state.State{Simulador: &state.ProcessInfo{
		PID:        99999999,
		Porta:      state.PortaPadraoSimulador,
		IniciadoEm: time.Now().UTC(),
	}}
	require.NoError(t, st.Save())
	r, err := Parar(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "stale_pid", r.Metodo)

	// state.json deve estar limpo.
	st2, err := state.Load()
	require.NoError(t, err)
	assert.Nil(t, st2.Simulador)
}

func TestEscolherPorta_ForcadaIndisponivel(t *testing.T) {
	// ocupar uma porta com httptest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	p := portaDe(t, srv.URL)
	_, err := escolherPorta(OpcoesIniciar{PortaForcada: p})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "em uso")
}

func TestEscolherPorta_AutoDetectaProxima(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	ocupada := portaDe(t, srv.URL)
	p, err := escolherPorta(OpcoesIniciar{PortaPreferida: ocupada})
	require.NoError(t, err)
	assert.NotEqual(t, ocupada, p)
}
