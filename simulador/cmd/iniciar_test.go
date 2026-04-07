package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIniciarCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"iniciar"})
	assert.NoError(t, err)
	assert.Equal(t, "iniciar", cmd.Name())
}

func TestIniciarCommand_HasExpectedFlags(t *testing.T) {
	flags := []string{"porta", "source", "versao-simulador"}
	for _, f := range flags {
		assert.NotNil(t, iniciarCmd.Flags().Lookup(f), "flag %q deveria existir", f)
	}
}
