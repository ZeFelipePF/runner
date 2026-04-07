package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_DefaultValues(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 8088, cfg.Assinador.PortaPadrao)
	assert.Equal(t, 9090, cfg.Simulador.PortaPadrao)
	assert.Equal(t, 21, cfg.JDK.VersaoMinima)
	assert.Equal(t, "temurin", cfg.JDK.Distribuicao)
}

func TestConfig_LoadMissing_ReturnsDefaults(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)

	_, err := EnsureDir()
	require.NoError(t, err)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 8088, cfg.Assinador.PortaPadrao)
}

func TestConfig_LoadPartial_FillsDefaults(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)

	dir, err := EnsureDir()
	require.NoError(t, err)

	partial := map[string]any{
		"assinador": map[string]any{"portaPadrao": 9999},
	}
	data, _ := json.Marshal(partial)
	require.NoError(t, os.WriteFile(filepath.Join(dir, configFile), data, 0644))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 9999, cfg.Assinador.PortaPadrao)
	assert.Equal(t, 9090, cfg.Simulador.PortaPadrao)
	assert.Equal(t, 21, cfg.JDK.VersaoMinima)
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)

	_, err := EnsureDir()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Assinador.PortaPadrao = 7777
	require.NoError(t, cfg.SaveConfig())

	loaded, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 7777, loaded.Assinador.PortaPadrao)
}
