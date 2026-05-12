package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hubsaude/runner/assinatura/internal/assinador"
	"github.com/stretchr/testify/assert"
)

func TestCodigoSaida_MapeiaCodigoAssinador(t *testing.T) {
	casos := map[string]int{
		"PARAM_AUSENTE":           ExitParamAusente,
		"PARAM_INVALIDO":          ExitParamInvalido,
		"ALGORITMO_NAO_SUPORTADO": ExitAlgoritmo,
		"ASSINATURA_INVALIDA":     ExitAssinaturaInval,
		"DISPOSITIVO_INDISPONIVEL": ExitDispositivo,
		"PIN_INVALIDO":            ExitDispositivo,
		"ERRO_INTERNO":            ExitInterno,
		"DESCONHECIDO":            ExitGenerico,
	}
	for codigo, esperado := range casos {
		err := &assinador.RespostaErro{Codigo: codigo, Mensagem: "x"}
		got := CodigoSaida(err)
		assert.Equalf(t, esperado, got, "codigo %s -> exit %d", codigo, esperado)
	}
}

func TestCodigoSaida_NilEGenerico(t *testing.T) {
	assert.Equal(t, ExitOK, CodigoSaida(nil))
	assert.Equal(t, ExitGenerico, CodigoSaida(errors.New("qualquer outro")))
}

func TestCodigoSaida_RespostaErroEncadeada(t *testing.T) {
	wrapped := fmt.Errorf("ao chamar /sign: %w", &assinador.RespostaErro{Codigo: "PARAM_AUSENTE"})
	assert.Equal(t, ExitParamAusente, CodigoSaida(wrapped))
}

func TestEmitirErro_RespostaErro_FormatoJSON(t *testing.T) {
	var buf bytes.Buffer
	EmitirErro(&buf, &assinador.RespostaErro{Codigo: "PARAM_INVALIDO", Mensagem: "campo X invalido"})
	var out map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "PARAM_INVALIDO", out["error"])
	assert.Equal(t, "campo X invalido", out["message"])
}

func TestEmitirErro_ErroGenerico(t *testing.T) {
	var buf bytes.Buffer
	EmitirErro(&buf, errors.New("falha qualquer"))
	var out map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "CLI_ERROR", out["error"])
	assert.Equal(t, "falha qualquer", out["message"])
}
