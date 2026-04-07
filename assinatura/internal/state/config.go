package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFile = "config.json"

// AssinadorConfig armazena configuracoes do assinador.
type AssinadorConfig struct {
	PortaPadrao             int `json:"portaPadrao"`
	TimeoutShutdownSegundos int `json:"timeoutShutdownSegundos"`
}

// SimuladorConfig armazena configuracoes do simulador.
type SimuladorConfig struct {
	PortaPadrao int    `json:"portaPadrao"`
	SourceURL   string `json:"sourceUrl"`
}

// JDKConfig armazena configuracoes do JDK.
type JDKConfig struct {
	VersaoMinima int    `json:"versaoMinima"`
	Distribuicao string `json:"distribuicao"`
}

// Config representa o conteudo de config.json.
type Config struct {
	Assinador AssinadorConfig `json:"assinador"`
	Simulador SimuladorConfig `json:"simulador"`
	JDK       JDKConfig       `json:"jdk"`
}

// DefaultConfig retorna a configuracao padrao.
func DefaultConfig() Config {
	return Config{
		Assinador: AssinadorConfig{
			PortaPadrao:             8088,
			TimeoutShutdownSegundos: 30,
		},
		Simulador: SimuladorConfig{
			PortaPadrao: 9090,
			SourceURL:   "",
		},
		JDK: JDKConfig{
			VersaoMinima: 21,
			Distribuicao: "temurin",
		},
	}
}

// LoadConfig le config.json, usando defaults para valores ausentes.
func LoadConfig() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	defaults := DefaultConfig()
	path := filepath.Join(dir, configFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &defaults, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ao ler %s: %w", configFile, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ao decodificar %s: %w", configFile, err)
	}

	if cfg.Assinador.PortaPadrao == 0 {
		cfg.Assinador.PortaPadrao = defaults.Assinador.PortaPadrao
	}
	if cfg.Assinador.TimeoutShutdownSegundos == 0 {
		cfg.Assinador.TimeoutShutdownSegundos = defaults.Assinador.TimeoutShutdownSegundos
	}
	if cfg.Simulador.PortaPadrao == 0 {
		cfg.Simulador.PortaPadrao = defaults.Simulador.PortaPadrao
	}
	if cfg.JDK.VersaoMinima == 0 {
		cfg.JDK.VersaoMinima = defaults.JDK.VersaoMinima
	}
	if cfg.JDK.Distribuicao == "" {
		cfg.JDK.Distribuicao = defaults.JDK.Distribuicao
	}

	return &cfg, nil
}

// SaveConfig escreve config.json.
func (c *Config) SaveConfig() error {
	dir, err := EnsureDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("ao codificar %s: %w", configFile, err)
	}

	path := filepath.Join(dir, configFile)
	return os.WriteFile(path, data, 0644)
}
