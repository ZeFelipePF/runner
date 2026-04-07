package state

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const hubsaudeDir = ".hubsaude"

// Dir retorna o caminho para ~/.hubsaude/ (ou %APPDATA%\.hubsaude\ no Windows).
func Dir() (string, error) {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", fmt.Errorf("ao determinar diretorio de estado: APPDATA nao definido")
		}
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("ao determinar diretorio de estado: %w", err)
		}
		base = home
	}
	return filepath.Join(base, hubsaudeDir), nil
}

// EnsureDir cria ~/.hubsaude/ e subdiretorios necessarios se nao existirem.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	subdirs := []string{
		"",
		"jdk",
		"simulador",
		"assinador",
	}

	for _, sub := range subdirs {
		p := filepath.Join(dir, sub)
		if err := os.MkdirAll(p, 0755); err != nil {
			return "", fmt.Errorf("ao criar diretorio %s: %w", p, err)
		}
	}

	return dir, nil
}
