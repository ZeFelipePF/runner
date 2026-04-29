package assinador

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ClienteHTTP fala com o assinador.jar no modo servidor.
type ClienteHTTP struct {
	BaseURL string
	HTTP    *http.Client
}

// NovoClienteHTTP cria um cliente apontando para http://127.0.0.1:porta com
// timeout padrao de 30s.
func NovoClienteHTTP(porta int) *ClienteHTTP {
	return &ClienteHTTP{
		BaseURL: "http://127.0.0.1:" + strconv.Itoa(porta),
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Executar envia POST /sign ou POST /validate com payloadJSON no corpo.
func (c *ClienteHTTP) Executar(ctx context.Context, op Operacao, payloadJSON []byte) (Resposta, error) {
	caminho := caminhoOperacao(op)
	if caminho == "" {
		return nil, fmt.Errorf("operacao desconhecida: %q", op)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+caminho,
		bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("ao montar requisicao %s: %w", caminho, err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.executar(req)
}

// Health consulta GET /health. Retorna nil se o servidor responde 2xx.
func (c *ClienteHTTP) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("/health retornou %d", resp.StatusCode)
	}
	return nil
}

// AguardarPronto faz polling em /health ate timeout. Retorna nil quando o servidor responde.
func (c *ClienteHTTP) AguardarPronto(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		ctxTent, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		err := c.Health(ctxTent)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("assinador nao respondeu em %s", timeout)
}

// Shutdown solicita encerramento do servidor.
func (c *ClienteHTTP) Shutdown(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/shutdown", nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("/shutdown retornou %d", resp.StatusCode)
	}
	return nil
}

func (c *ClienteHTTP) executar(req *http.Request) (Resposta, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ao chamar %s: %w", req.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ao ler resposta de %s: %w", req.URL, err)
	}

	if resp.StatusCode/100 != 2 {
		var re RespostaErro
		if json.Unmarshal(body, &re) == nil && re.Codigo != "" {
			return nil, &re
		}
		return nil, fmt.Errorf("%s retornou HTTP %d: %s", req.URL, resp.StatusCode, bytes.TrimSpace(body))
	}

	var r Resposta
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("ao decodificar resposta de %s: %w", req.URL, err)
	}
	return r, nil
}

func caminhoOperacao(op Operacao) string {
	switch op {
	case OperacaoCriar:
		return "/sign"
	case OperacaoValidar:
		return "/validate"
	}
	return ""
}
