package assinador

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/hubsaude/runner/assinatura/internal/state"
)

// LocalizarJar procura o assinador.jar em locais conhecidos, em ordem:
//  1. variavel de ambiente HUBSAUDE_ASSINADOR_JAR
//  2. ~/.hubsaude/assinador/assinador.jar
//  3. ./assinador.jar (cwd)
//  4. ./assinador/target/assinador.jar (layout dev)
func LocalizarJar() (string, error) {
	candidatos := []string{}

	if v := os.Getenv("HUBSAUDE_ASSINADOR_JAR"); v != "" {
		candidatos = append(candidatos, v)
	}

	if dir, err := state.Dir(); err == nil {
		candidatos = append(candidatos, filepath.Join(dir, "assinador", "assinador.jar"))
	}

	if cwd, err := os.Getwd(); err == nil {
		candidatos = append(candidatos,
			filepath.Join(cwd, "assinador.jar"),
			filepath.Join(cwd, "assinador", "target", "assinador.jar"),
		)
	}

	for _, c := range candidatos {
		if c == "" {
			continue
		}
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c, nil
		}
	}
	return "", errors.New("assinador.jar nao localizado (defina HUBSAUDE_ASSINADOR_JAR ou copie o jar para ~/.hubsaude/assinador/)")
}
