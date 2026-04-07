package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPararCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"parar"})
	assert.NoError(t, err)
	assert.Equal(t, "parar", cmd.Name())
}
