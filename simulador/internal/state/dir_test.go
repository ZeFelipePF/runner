package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDir_ReturnsHubsaudePath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("APPDATA", tmp)

	dir, err := Dir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, ".hubsaude"), dir)
}

func TestEnsureDir_CreatesSubdirectories(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("APPDATA", tmp)

	dir, err := EnsureDir()
	require.NoError(t, err)

	expected := []string{"jdk", "simulador", "assinador"}
	for _, sub := range expected {
		info, err := os.Stat(filepath.Join(dir, sub))
		assert.NoError(t, err, "subdiretorio %q deveria existir", sub)
		assert.True(t, info.IsDir())
	}
}

func TestEnsureDir_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("APPDATA", tmp)

	dir1, err := EnsureDir()
	require.NoError(t, err)
	dir2, err := EnsureDir()
	require.NoError(t, err)
	assert.Equal(t, dir1, dir2)
}
