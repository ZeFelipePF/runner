package assinador

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

// ClienteCLI invoca o assinador.jar diretamente via "java -jar".
type ClienteCLI struct {
	JavaPath string // caminho do binario java (default: "java" no PATH)
	JarPath  string // caminho absoluto do assinador.jar
}

// NovoClienteCLI cria um cliente com defaults razoaveis.
func NovoClienteCLI(javaPath, jarPath string) *ClienteCLI {
	if javaPath == "" {
		javaPath = "java"
	}
	return &ClienteCLI{JavaPath: javaPath, JarPath: jarPath}
}

// Executar invoca "java -jar <jar> <op> --input -" enviando payloadJSON via stdin.
// Retorna o JSON decodificado em caso de sucesso, ou *RespostaErro encapsulado em error.
func (c *ClienteCLI) Executar(ctx context.Context, op Operacao, payloadJSON []byte) (Resposta, error) {
	if c.JarPath == "" {
		return nil, errors.New("ClienteCLI: caminho do assinador.jar nao definido")
	}
	cmd := exec.CommandContext(ctx, c.JavaPath, "-jar", c.JarPath, string(op), "--input", "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(payloadJSON)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Tenta extrair RespostaErro do stderr antes de propagar erro de processo.
		if respErro := decodificarErro(stderr.Bytes()); respErro != nil {
			return nil, respErro
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("assinador.jar saiu com codigo %d: %s",
				exitErr.ExitCode(), bytes.TrimSpace(stderr.Bytes()))
		}
		return nil, fmt.Errorf("ao executar %s: %w", c.JavaPath, err)
	}

	if stdout.Len() == 0 {
		return nil, fmt.Errorf("assinador.jar retornou stdout vazio (stderr=%s)",
			bytes.TrimSpace(stderr.Bytes()))
	}

	var resp Resposta
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("ao decodificar resposta do assinador.jar: %w", err)
	}
	return resp, nil
}

func decodificarErro(stderr []byte) *RespostaErro {
	stderr = bytes.TrimSpace(stderr)
	if len(stderr) == 0 {
		return nil
	}
	// stderr pode conter linhas de log JSON antes do payload de erro; pega a ultima linha JSON.
	linhas := bytes.Split(stderr, []byte("\n"))
	for i := len(linhas) - 1; i >= 0; i-- {
		linha := bytes.TrimSpace(linhas[i])
		if len(linha) == 0 || linha[0] != '{' {
			continue
		}
		var r RespostaErro
		if err := json.Unmarshal(linha, &r); err == nil && r.Codigo != "" {
			return &r
		}
	}
	return nil
}
