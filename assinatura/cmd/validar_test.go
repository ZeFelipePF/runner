package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidarCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"validar"})
	assert.NoError(t, err)
	assert.Equal(t, "validar", cmd.Name())
}

func TestValidarCommand_HasExpectedFlags(t *testing.T) {
	flags := []string{"message-file", "signature", "public-key", "modo", "porta"}
	for _, f := range flags {
		assert.NotNil(t, validarCmd.Flags().Lookup(f), "flag %q deveria existir", f)
	}
}
