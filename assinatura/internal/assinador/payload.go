// Package assinador encapsula a invocacao do assinador.jar nos modos local e HTTP.
package assinador

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Operacao identifica a acao a executar (criar/validar).
type Operacao string

const (
	OperacaoCriar   Operacao = "sign"
	OperacaoValidar Operacao = "validate"
)

// RespostaErro replica o formato JSON retornado pelo assinador em falhas.
type RespostaErro struct {
	Codigo   string `json:"error"`
	Mensagem string `json:"message"`
}

func (e *RespostaErro) Error() string {
	if e == nil {
		return ""
	}
	if e.Mensagem == "" {
		return e.Codigo
	}
	return fmt.Sprintf("%s: %s", e.Codigo, e.Mensagem)
}

// Resposta representa a resposta JSON do assinador (estrutura aberta).
type Resposta map[string]any

// LerPayload le um JSON arbitrario de arquivo (ou stdin se caminho == "" ou "-").
// Retorna o map decodificado e os bytes brutos (para reenvio sem modificacoes).
func LerPayload(caminho string) (Resposta, []byte, error) {
	bytes, err := lerBytes(caminho)
	if err != nil {
		return nil, nil, err
	}
	if len(bytes) == 0 {
		return nil, nil, fmt.Errorf("payload vazio")
	}
	var p Resposta
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, nil, fmt.Errorf("ao decodificar payload JSON: %w", err)
	}
	return p, bytes, nil
}

func lerBytes(caminho string) ([]byte, error) {
	if caminho == "" || caminho == "-" {
		return io.ReadAll(os.Stdin)
	}
	info, err := os.Stat(caminho)
	if err != nil {
		return nil, fmt.Errorf("ao abrir payload %q: %w", caminho, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("payload %q e um diretorio", caminho)
	}
	return os.ReadFile(caminho)
}
