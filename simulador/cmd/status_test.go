package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"status"})
	assert.NoError(t, err)
	assert.Equal(t, "status", cmd.Name())
}
