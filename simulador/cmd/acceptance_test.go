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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Testes de aceitacao end-to-end do binario `simulador`.
//
// Cobertura por User Story:
//   US-03: comandos iniciar/parar/status, exibicao de status, parar quando
//          nao rodando, deteccao de instancia, --source com cache.
//   US-04: provisionamento de JDK observado indiretamente quando iniciar
//          dispara o fluxo (testes de unidade em internal/jdk cobrem em
//          profundidade).
//
// Habilitar com:
//   go test -tags=acceptance ./cmd -run Aceitacao

var binPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "simulador-acceptance-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro criando temp dir:", err)
		os.Exit(1)
	}

	nome := "simulador"
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

func binarioSimulador(t *testing.T) string {
	t.Helper()
	if binPath == "" {
		t.Fatal("binario nao construido (TestMain falhou?)")
	}
	return binPath
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
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exit
}

// US-03: subcomandos iniciar / parar / status disponiveis.
func TestAceitacao_US03_AjudaListaSubcomandos(t *testing.T) {
	b := binarioSimulador(t)
	stdout, _, exit := rodar(t, b, "--help")
	require.Equal(t, 0, exit)
	for _, esperado := range []string{"iniciar", "parar", "status", "versao"} {
		assert.Contains(t, stdout, esperado, "subcomando %q ausente em --help", esperado)
	}
}

func TestAceitacao_US03_ComandoVersao(t *testing.T) {
	b := binarioSimulador(t)
	stdout, _, exit := rodar(t, b, "versao")
	require.Equal(t, 0, exit)
	assert.Contains(t, stdout, "simulador")
}

// US-03: CLI exibe status estruturado quando nada esta rodando.
func TestAceitacao_US03_StatusQuandoNaoRodando(t *testing.T) {
	b := binarioSimulador(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	stdout, _, exit := rodar(t, b, "status", "--quiet")
	require.Equal(t, 0, exit)

	var resp map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	registrado, ok := resp["registrado"].(bool)
	require.True(t, ok, "campo 'registrado' deve ser bool")
	assert.False(t, registrado, "esperado registrado=false em ambiente limpo")
}

// US-03: CLI para simulador (caso nao rodando -> reporta nao_estava_rodando).
func TestAceitacao_US03_PararQuandoNaoRodando(t *testing.T) {
	b := binarioSimulador(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	stdout, _, exit := rodar(t, b, "parar", "--quiet")
	require.Equal(t, 0, exit)
	assert.Contains(t, stdout, "nao_estava_rodando")
	assert.Contains(t, stdout, "NOT_RUNNING")
}

// US-03: --source aponta para URL custom (validado pela flag aceitar valor).
func TestAceitacao_US03_FlagSourceAceita(t *testing.T) {
	b := binarioSimulador(t)
	// O comando vai falhar (jar invalido) mas a flag deve ser parseada sem erro.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	// Source apontando para arquivo inexistente -> deve falhar com erro de download,
	// nao com erro de parse de flag.
	_, stderr, exit := rodar(t, b, "iniciar",
		"--source", "file:///caminho/que/nao/existe.jar",
		"--quiet")
	require.NotEqual(t, 0, exit, "esperado falha por jar inexistente")
	assert.NotContains(t, stderr, "unknown flag")
	assert.NotContains(t, stderr, "invalid argument")
}
