package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := rootCmd
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "simulador")
}

func TestRootCommand_HasExpectedSubcommands(t *testing.T) {
	expected := []string{"iniciar", "parar", "status", "versao"}
	names := make([]string, 0)
	for _, c := range rootCmd.Commands() {
		names = append(names, c.Name())
	}
	for _, e := range expected {
		assert.Contains(t, names, e, "subcomando %q deveria existir", e)
	}
}
