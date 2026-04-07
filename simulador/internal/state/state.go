package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const stateFile = "state.json"

// ProcessInfo representa um processo Java em execucao.
type ProcessInfo struct {
	PID        int       `json:"pid"`
	Porta      int       `json:"porta"`
	IniciadoEm time.Time `json:"iniciadoEm"`
	JavaPath   string    `json:"javaPath"`
	Versao     string    `json:"versao,omitempty"`
}

// State representa o conteudo de state.json.
type State struct {
	Assinador *ProcessInfo `json:"assinador,omitempty"`
	Simulador *ProcessInfo `json:"simulador,omitempty"`
}

// Load le state.json do diretorio hubsaude.
// Retorna um State vazio se o arquivo nao existir.
func Load() (*State, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, stateFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ao ler %s: %w", stateFile, err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("ao decodificar %s: %w", stateFile, err)
	}
	return &s, nil
}

// Save escreve state.json no diretorio hubsaude.
func (s *State) Save() error {
	dir, err := EnsureDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("ao codificar %s: %w", stateFile, err)
	}

	path := filepath.Join(dir, stateFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("ao escrever %s: %w", stateFile, err)
	}
	return nil
}

// IsProcessAlive verifica se um PID ainda esta em execucao.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return isProcessAlive(pid)
}

// CleanStale remove entradas de state.json onde o processo nao existe mais.
// Retorna true se houve alteracao.
func (s *State) CleanStale() bool {
	changed := false
	if s.Assinador != nil && !IsProcessAlive(s.Assinador.PID) {
		s.Assinador = nil
		changed = true
	}
	if s.Simulador != nil && !IsProcessAlive(s.Simulador.PID) {
		s.Simulador = nil
		changed = true
	}
	return changed
}
