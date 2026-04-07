package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServidorSubcommands_Registered(t *testing.T) {
	subcommands := []string{"iniciar", "parar", "status"}
	for _, name := range subcommands {
		cmd, _, err := servidorCmd.Find([]string{name})
		assert.NoError(t, err)
		assert.Equal(t, name, cmd.Name())
	}
}

func TestServidorIniciar_HasPortaFlag(t *testing.T) {
	assert.NotNil(t, servidorIniciarCmd.Flags().Lookup("porta"))
}

func TestServidorIniciar_HasTimeoutFlag(t *testing.T) {
	assert.NotNil(t, servidorIniciarCmd.Flags().Lookup("timeout"))
}
