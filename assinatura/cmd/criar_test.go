package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCriarCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"criar"})
	assert.NoError(t, err)
	assert.Equal(t, "criar", cmd.Name())
}

func TestCriarCommand_HasExpectedFlags(t *testing.T) {
	flags := []string{"payload", "modo", "porta", "jar"}
	for _, f := range flags {
		assert.NotNil(t, criarCmd.Flags().Lookup(f), "flag %q deveria existir", f)
	}
}
