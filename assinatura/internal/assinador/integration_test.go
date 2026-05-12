//go:build integration

package assinador

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Testes de integracao end-to-end CLI <-> assinador.jar real.
//
// Habilitar com:
//   go test -tags=integration ./internal/assinador -run Integracao
//
// Requer:
//   - JDK 21+ no PATH (ou JAVA_HOME apontando)
//   - assinador.jar disponivel:
//       * env HUBSAUDE_ASSINADOR_JAR=<caminho>, OU
//       * runner/assinador/target/assinador.jar (resolvido automaticamente)

func jarPath(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("HUBSAUDE_ASSINADOR_JAR"); p != "" {
		return p
	}
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// .../assinatura/internal/assinador -> sobe 3 niveis ate o repo root.
	root := filepath.Clean(filepath.Join(cwd, "..", "..", ".."))
	candidato := filepath.Join(root, "assinador", "target", "assinador.jar")
	if _, err := os.Stat(candidato); err == nil {
		return candidato
	}
	t.Skipf("assinador.jar nao encontrado (procurado em %s; defina HUBSAUDE_ASSINADOR_JAR)", candidato)
	return ""
}

func payloadCriacaoValido() []byte {
	p := map[string]any{
		"bundle":     `{"resourceType":"Bundle"}`,
		"provenance": `{"resourceType":"Provenance"}`,
		"materialCriptografico": map[string]any{
			"tipo":         "PEM",
			"chavePrivada": "----PEM----",
		},
		"certificados":         []string{"YWJj"},
		"timestampReferencia":  1700000000,
		"estrategiaTimestamp":  "iat",
		"politicaAssinatura":   "urn:test:1.0",
	}
	b, _ := json.Marshal(p)
	return b
}

func payloadValidacaoValido() []byte {
	hash64 := ""
	for i := 0; i < 64; i++ {
		hash64 += "a"
	}
	p := map[string]any{
		"jws":                "YWJj",
		"trustStore":         []string{hash64},
		"minCertIssueDate":   0,
		"referenceTimestamp": 0,
		"signaturePolicyId":  "urn:test:1.0",
	}
	b, _ := json.Marshal(p)
	return b
}

func TestIntegracaoModoLocal_Sign(t *testing.T) {
	jar := jarPath(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := NovoClienteCLI("", jar)
	resp, err := c.Executar(ctx, OperacaoCriar, payloadCriacaoValido())
	require.NoError(t, err)
	assert.Equal(t, "Signature", resp["resourceType"])
}

func TestIntegracaoModoLocal_Validate(t *testing.T) {
	jar := jarPath(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := NovoClienteCLI("", jar)
	resp, err := c.Executar(ctx, OperacaoValidar, payloadValidacaoValido())
	require.NoError(t, err)
	assert.Equal(t, "OperationOutcome", resp["resourceType"])
}

func TestIntegracaoModoLocal_PayloadInvalido(t *testing.T) {
	jar := jarPath(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := NovoClienteCLI("", jar)
	_, err := c.Executar(ctx, OperacaoCriar, []byte("{}"))
	require.Error(t, err)
	if re, ok := err.(*RespostaErro); ok {
		assert.Equal(t, "PARAM_AUSENTE", re.Codigo)
	}
}

func TestIntegracaoModoHTTP_SignViaGarantir(t *testing.T) {
	jar := jarPath(t)
	redirHome(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	srv, err := Garantir(ctx, OpcoesStartup{
		JarPath:        jar,
		PortaPreferida: 18088,
		TimeoutPronto:  30 * time.Second,
	})
	require.NoError(t, err)
	defer func() {
		_ = NovoClienteHTTP(srv.Porta).Shutdown(ctx)
	}()

	cliente := NovoClienteHTTP(srv.Porta)
	resp, err := cliente.Executar(ctx, OperacaoCriar, payloadCriacaoValido())
	require.NoError(t, err)
	assert.Equal(t, "Signature", resp["resourceType"])
}

func TestIntegracaoModoHTTP_Reuso(t *testing.T) {
	jar := jarPath(t)
	redirHome(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	srv1, err := Garantir(ctx, OpcoesStartup{
		JarPath:        jar,
		PortaPreferida: 18099,
		TimeoutPronto:  30 * time.Second,
	})
	require.NoError(t, err)
	defer func() {
		_ = NovoClienteHTTP(srv1.Porta).Shutdown(ctx)
	}()

	srv2, err := Garantir(ctx, OpcoesStartup{
		JarPath:        jar,
		PortaPreferida: 18099,
		TimeoutPronto:  5 * time.Second,
	})
	require.NoError(t, err)
	assert.True(t, srv2.Reusado, "segunda chamada deve reusar instancia")
	assert.Equal(t, srv1.Porta, srv2.Porta)
	assert.Equal(t, srv1.PID, srv2.PID)
}

// Garante que `java` esteja no PATH quando executando integracao.
func TestMain(m *testing.M) {
	if _, err := exec.LookPath("java"); err != nil {
		os.Stderr.WriteString("[skip] java nao encontrado no PATH; testes de integracao serao pulados\n")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func redirHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("APPDATA", dir)
}
