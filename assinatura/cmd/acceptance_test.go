//go:build acceptance

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Testes de aceitacao end-to-end do binario `assinatura`.
//
// Cobertura por User Story:
//   US-01: comandos criar/validar/servidor, modos local e HTTP, formatacao
//          de saida, codigos de saida, reuso de instancia.
//   US-02: o binario assinador.jar valida parametros e simula assinatura,
//          observado atraves da CLI.
//
// Habilitar com:
//   go test -tags=acceptance ./cmd -run Aceitacao
//
// Requisitos opcionais:
//   - JDK 21+ no PATH (testes que precisam do jar sao pulados sem java)
//   - HUBSAUDE_ASSINADOR_JAR ou ../assinador/target/assinador.jar para
//     cenarios end-to-end com a jar real.

var binPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "assinatura-acceptance-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro criando temp dir:", err)
		os.Exit(1)
	}

	nome := "assinatura"
	if runtime.GOOS == "windows" {
		nome += ".exe"
	}
	binPath = filepath.Join(tmp, nome)

	cwd, _ := os.Getwd()
	modDir := filepath.Clean(filepath.Join(cwd, ".."))
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = modDir
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build falhou: %v\n%s\n", err, out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func binarioAssinatura(t *testing.T) string {
	t.Helper()
	if binPath == "" {
		t.Fatal("binario nao construido (TestMain falhou?)")
	}
	return binPath
}

func acharJarAssinador(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("HUBSAUDE_ASSINADOR_JAR"); p != "" {
		return p
	}
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// .../assinatura/cmd -> 2 niveis ate o root.
	root := filepath.Clean(filepath.Join(cwd, "..", ".."))
	candidato := filepath.Join(root, "assinador", "target", "assinador.jar")
	if _, err := os.Stat(candidato); err == nil {
		return candidato
	}
	t.Skipf("assinador.jar nao encontrado (esperado em %s)", candidato)
	return ""
}

func rodar(t *testing.T, b string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(b, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exit = exitErr.ExitCode()
		} else {
			t.Fatalf("erro inesperado executando binario: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exit
}

func rodarComStdin(t *testing.T, b string, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(b, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exit = exitErr.ExitCode()
		} else {
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exit
}

// --- US-01: CLI aceita comandos criar e validar -----------------------------

func TestAceitacao_US01_AjudaListaSubcomandos(t *testing.T) {
	b := binarioAssinatura(t)
	stdout, _, exit := rodar(t, b, "--help")
	require.Equal(t, 0, exit)
	for _, esperado := range []string{"criar", "validar", "servidor", "versao"} {
		assert.Contains(t, stdout, esperado, "subcomando %q deve aparecer em --help", esperado)
	}
}

func TestAceitacao_US01_ComandoVersao(t *testing.T) {
	b := binarioAssinatura(t)
	stdout, _, exit := rodar(t, b, "versao")
	require.Equal(t, 0, exit)
	assert.Contains(t, stdout, "assinatura")
}

// US-01 + US-02: CLI invoca assinador.jar com parametros e exibe resposta legivel.
func TestAceitacao_US01_CriarModoLocal_PayloadValido(t *testing.T) {
	b := binarioAssinatura(t)
	jar := acharJarAssinador(t)
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java nao encontrado no PATH")
	}

	payload := payloadCriacao(t)
	stdout, _, exit := rodarComStdin(t, b, payload,
		"criar", "--modo", "local", "--jar", jar, "--payload", "-", "--quiet")
	require.Equal(t, 0, exit, "stdout=%s", stdout)

	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	assert.Equal(t, "Signature", resp["resourceType"])
}

// US-02: validacao de parametros retorna mensagens de erro claras + codigo de saida.
func TestAceitacao_US02_PayloadInvalidoRetornaErroEstruturado(t *testing.T) {
	b := binarioAssinatura(t)
	jar := acharJarAssinador(t)
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java nao encontrado no PATH")
	}

	stdout, stderr, exit := rodarComStdin(t, b, "{}",
		"criar", "--modo", "local", "--jar", jar, "--payload", "-", "--quiet")
	require.NotEqual(t, 0, exit, "esperado exit nao-zero para payload invalido")
	assert.Empty(t, stdout, "stdout deve estar vazio em caso de erro (saida JSON vai pra stderr)")
	assert.Contains(t, stderr, "PARAM_AUSENTE", "stderr deve conter codigo de erro estruturado")
}

func TestAceitacao_US01_ValidarModoLocal(t *testing.T) {
	b := binarioAssinatura(t)
	jar := acharJarAssinador(t)
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java nao encontrado no PATH")
	}

	payload := payloadValidacao(t)
	stdout, _, exit := rodarComStdin(t, b, payload,
		"validar", "--modo", "local", "--jar", jar, "--payload", "-", "--quiet")
	require.Equal(t, 0, exit, "stdout=%s", stdout)

	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	assert.Equal(t, "OperationOutcome", resp["resourceType"])
}

// US-01: servidor status quando nada esta rodando -> retorna JSON estruturado.
func TestAceitacao_US01_ServidorStatusQuandoNaoRodando(t *testing.T) {
	b := binarioAssinatura(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	stdout, _, exit := rodar(t, b, "servidor", "status", "--quiet")
	require.Equal(t, 0, exit)

	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	running, ok := resp["running"].(bool)
	require.True(t, ok, "campo 'running' deve ser bool, resp=%v", resp)
	assert.False(t, running, "running=false esperado em ambiente limpo")
}

// US-01: servidor parar quando nada esta rodando -> 0 + status NOT_RUNNING.
func TestAceitacao_US01_ServidorPararQuandoNaoRodando(t *testing.T) {
	b := binarioAssinatura(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	stdout, _, exit := rodar(t, b, "servidor", "parar", "--quiet")
	require.Equal(t, 0, exit)
	assert.Contains(t, stdout, "NOT_RUNNING")
}

// US-01: codigo de saida para arquivo de payload inexistente.
func TestAceitacao_US01_PayloadInexistenteRetornaErro(t *testing.T) {
	b := binarioAssinatura(t)
	_, _, exit := rodar(t, b,
		"criar", "--modo", "local", "--payload", "/caminho/que/nao/existe.json", "--quiet")
	require.NotEqual(t, 0, exit)
}

// Helpers de payload, replicados dos testes de integracao.

func payloadCriacao(t *testing.T) string {
	t.Helper()
	p := map[string]any{
		"bundle":     `{"resourceType":"Bundle"}`,
		"provenance": `{"resourceType":"Provenance"}`,
		"materialCriptografico": map[string]any{
			"tipo":         "PEM",
			"chavePrivada": "----PEM----",
		},
		"certificados":        []string{"YWJj"},
		"timestampReferencia": 1700000000,
		"estrategiaTimestamp": "iat",
		"politicaAssinatura":  "urn:test:1.0",
	}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	return string(b)
}

func payloadValidacao(t *testing.T) string {
	t.Helper()
	hash64 := strings.Repeat("a", 64)
	p := map[string]any{
		"jws":                "YWJj",
		"trustStore":         []string{hash64},
		"minCertIssueDate":   0,
		"referenceTimestamp": 0,
		"signaturePolicyId":  "urn:test:1.0",
	}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	return string(b)
}
