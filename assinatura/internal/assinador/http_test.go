package assinador

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClienteHTTP_ExecutarSucesso(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sign", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"k":"v"}`, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resourceType":"Signature","data":"abc"}`))
	}))
	defer srv.Close()

	c := clienteApontandoPara(t, srv.URL)
	r, err := c.Executar(context.Background(), OperacaoCriar, []byte(`{"k":"v"}`))
	require.NoError(t, err)
	assert.Equal(t, "Signature", r["resourceType"])
}

func TestClienteHTTP_ErroEstruturado(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"PARAM_AUSENTE","message":"campo X"}`))
	}))
	defer srv.Close()

	c := clienteApontandoPara(t, srv.URL)
	_, err := c.Executar(context.Background(), OperacaoValidar, []byte(`{}`))
	require.Error(t, err)
	var re *RespostaErro
	require.ErrorAs(t, err, &re)
	assert.Equal(t, "PARAM_AUSENTE", re.Codigo)
}

func TestClienteHTTP_HealthENaoOk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := clienteApontandoPara(t, srv.URL)
	err := c.Health(context.Background())
	assert.ErrorContains(t, err, "503")
}

func TestClienteHTTP_AguardarProntoTimeout(t *testing.T) {
	c := NovoClienteHTTP(1)
	c.BaseURL = "http://127.0.0.1:1"
	err := c.AguardarPronto(context.Background(), 200*time.Millisecond)
	assert.Error(t, err)
}

func TestClienteHTTP_OperacaoDesconhecida(t *testing.T) {
	c := NovoClienteHTTP(8088)
	_, err := c.Executar(context.Background(), Operacao("xyz"), []byte(`{}`))
	assert.ErrorContains(t, err, "operacao desconhecida")
}

func clienteApontandoPara(t *testing.T, baseURL string) *ClienteHTTP {
	t.Helper()
	host, porta := hostPorta(t, baseURL)
	c := NovoClienteHTTP(porta)
	c.BaseURL = "http://" + net.JoinHostPort(host, strconv.Itoa(porta))
	return c
}

func hostPorta(t *testing.T, url string) (string, int) {
	t.Helper()
	u := strings.TrimPrefix(url, "http://")
	host, p, err := net.SplitHostPort(u)
	require.NoError(t, err)
	porta, err := strconv.Atoi(p)
	require.NoError(t, err)
	return host, porta
}
