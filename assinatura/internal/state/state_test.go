package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_LoadEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)

	_, err := EnsureDir()
	require.NoError(t, err)

	s, err := Load()
	require.NoError(t, err)
	assert.Nil(t, s.Assinador)
	assert.Nil(t, s.Simulador)
}

func TestState_SaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)

	_, err := EnsureDir()
	require.NoError(t, err)

	s := &State{
		Assinador: &ProcessInfo{
			PID:        12345,
			Porta:      8088,
			IniciadoEm: time.Now().UTC().Truncate(time.Second),
			JavaPath:   "/usr/bin/java",
		},
	}
	require.NoError(t, s.Save())

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 12345, loaded.Assinador.PID)
	assert.Equal(t, 8088, loaded.Assinador.Porta)
}

func TestState_CleanStale_RemovesDeadProcess(t *testing.T) {
	s := &State{
		Assinador: &ProcessInfo{
			PID:   999999,
			Porta: 8088,
		},
	}
	changed := s.CleanStale()
	assert.True(t, changed)
	assert.Nil(t, s.Assinador)
}

func TestState_SaveCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)

	_, err := EnsureDir()
	require.NoError(t, err)

	s := &State{}
	require.NoError(t, s.Save())

	dir, _ := Dir()
	_, err = os.Stat(filepath.Join(dir, stateFile))
	assert.NoError(t, err)
}
