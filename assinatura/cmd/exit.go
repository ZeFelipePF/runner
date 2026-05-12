package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/hubsaude/runner/assinatura/internal/assinador"
)

// ExitCode mapeia codigos de erro do assinador para exit codes do CLI.
// Alinhado com br.gov.saude.assinador.erro.MapeadorErro (lado Java).
const (
	ExitOK              = 0
	ExitGenerico        = 1
	ExitParamAusente    = 2
	ExitParamInvalido   = 3
	ExitAlgoritmo       = 4
	ExitAssinaturaInval = 5
	ExitDispositivo     = 6
	ExitInterno         = 7
)

// CodigoSaida retorna o exit code apropriado para um erro propagado pelo CLI.
// Reconhece *assinador.RespostaErro e mapeia seu Codigo; demais erros => ExitGenerico.
func CodigoSaida(err error) int {
	if err == nil {
		return ExitOK
	}
	var re *assinador.RespostaErro
	if errors.As(err, &re) {
		switch re.Codigo {
		case "PARAM_AUSENTE":
			return ExitParamAusente
		case "PARAM_INVALIDO":
			return ExitParamInvalido
		case "ALGORITMO_NAO_SUPORTADO":
			return ExitAlgoritmo
		case "ASSINATURA_INVALIDA":
			return ExitAssinaturaInval
		case "DISPOSITIVO_INDISPONIVEL", "PIN_INVALIDO":
			return ExitDispositivo
		case "ERRO_INTERNO":
			return ExitInterno
		}
		return ExitGenerico
	}
	return ExitGenerico
}

// EmitirErro escreve uma representacao JSON estruturada do erro em stderr.
// *RespostaErro e serializado preservando codigo/mensagem; demais erros usam
// codigo "CLI_ERROR".
func EmitirErro(w io.Writer, err error) {
	if err == nil {
		return
	}
	var re *assinador.RespostaErro
	payload := map[string]any{}
	if errors.As(err, &re) {
		payload["error"] = re.Codigo
		payload["message"] = re.Mensagem
	} else {
		payload["error"] = "CLI_ERROR"
		payload["message"] = err.Error()
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}

// ExecuteOrExit roda a CLI e propaga exit codes estruturados.
func ExecuteOrExit() {
	if err := Execute(); err != nil {
		EmitirErro(os.Stderr, err)
		os.Exit(CodigoSaida(err))
	}
}
