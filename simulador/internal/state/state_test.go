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
		Simulador: &ProcessInfo{
			PID:        54321,
			Porta:      9090,
			IniciadoEm: time.Now().UTC().Truncate(time.Second),
			JavaPath:   "/usr/bin/java",
			Versao:     "0.0.1-SNAPSHOT",
		},
	}
	require.NoError(t, s.Save())

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 54321, loaded.Simulador.PID)
	assert.Equal(t, 9090, loaded.Simulador.Porta)
	assert.Equal(t, "0.0.1-SNAPSHOT", loaded.Simulador.Versao)
}

func TestState_CleanStale_RemovesDeadProcess(t *testing.T) {
	s := &State{
		Simulador: &ProcessInfo{
			PID:   999999,
			Porta: 9090,
		},
	}
	changed := s.CleanStale()
	assert.True(t, changed)
	assert.Nil(t, s.Simulador)
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
