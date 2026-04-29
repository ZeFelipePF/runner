package assinador

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClienteCLI_SemJarPath(t *testing.T) {
	c := NovoClienteCLI("", "")
	_, err := c.Executar(context.Background(), OperacaoCriar, []byte(`{}`))
	assert.ErrorContains(t, err, "assinador.jar nao definido")
}

func TestClienteCLI_StdoutSucesso(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("usa shell unix")
	}
	bin := scriptShell(t, `#!/bin/sh
echo '{"resourceType":"Signature","data":"abc"}'
`)
	c := NovoClienteCLI(bin, "/dev/null")
	resp, err := c.Executar(context.Background(), OperacaoCriar, []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "Signature", resp["resourceType"])
}

func TestClienteCLI_RespostaErroEstruturada(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("usa shell unix")
	}
	bin := scriptShell(t, `#!/bin/sh
echo '{"error":"PARAM_AUSENTE","message":"campo obrigatorio"}' 1>&2
exit 2
`)
	c := NovoClienteCLI(bin, "/dev/null")
	_, err := c.Executar(context.Background(), OperacaoCriar, []byte(`{}`))
	require.Error(t, err)
	var re *RespostaErro
	assert.ErrorAs(t, err, &re)
	assert.Equal(t, "PARAM_AUSENTE", re.Codigo)
}

func TestClienteCLI_ExitSemRespostaErro(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("usa shell unix")
	}
	bin := scriptShell(t, `#!/bin/sh
echo "erro nao estruturado" 1>&2
exit 7
`)
	c := NovoClienteCLI(bin, "/dev/null")
	_, err := c.Executar(context.Background(), OperacaoCriar, []byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "codigo 7")
}

func scriptShell(t *testing.T, conteudo string) string {
	t.Helper()
	dir := t.TempDir()
	caminho := filepath.Join(dir, "fake-java")
	require.NoError(t, os.WriteFile(caminho, []byte(conteudo), 0755))
	return caminho
}
