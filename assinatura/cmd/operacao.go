package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hubsaude/runner/assinatura/internal/assinador"
	"github.com/hubsaude/runner/assinatura/internal/state"
	"github.com/spf13/cobra"
)

// executarOperacaoAssinador encapsula o fluxo dos comandos `criar` e `validar`:
// le payload, escolhe modo (local|http), executa, imprime resposta no stdout.
func executarOperacaoAssinador(cmd *cobra.Command, op string) error {
	payloadPath, _ := cmd.Flags().GetString("payload")
	modo, _ := cmd.Flags().GetString("modo")
	porta, _ := cmd.Flags().GetInt("porta")
	jar, _ := cmd.Flags().GetString("jar")

	_, payloadBytes, err := assinador.LerPayload(payloadPath)
	if err != nil {
		return err
	}

	jarPath := jar
	if jarPath == "" {
		jarPath, err = assinador.LocalizarJar()
		if err != nil && modo != "http" {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
	defer cancel()

	var resp assinador.Resposta

	switch modo {
	case "local":
		cliente := assinador.NovoClienteCLI("", jarPath)
		resp, err = cliente.Executar(ctx, assinador.Operacao(op), payloadBytes)
	case "http", "":
		cfg, _ := state.LoadConfig()
		preferida := porta
		forcada := 0
		if preferida == 0 && cfg != nil {
			preferida = cfg.Assinador.PortaPadrao
		}
		if porta > 0 {
			forcada = porta
		}
		srv, gErr := assinador.Garantir(ctx, assinador.OpcoesStartup{
			JarPath:        jarPath,
			PortaPreferida: preferida,
			PortaForcada:   forcada,
			LogProgresso:   cmd.ErrOrStderr(),
		})
		if gErr != nil {
			return gErr
		}
		cliente := assinador.NovoClienteHTTP(srv.Porta)
		resp, err = cliente.Executar(ctx, assinador.Operacao(op), payloadBytes)
	default:
		return fmt.Errorf("modo desconhecido: %q (use 'local' ou 'http')", modo)
	}

	if err != nil {
		return err
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
}
