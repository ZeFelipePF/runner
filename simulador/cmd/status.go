package cmd

import (
	"context"
	"time"

	"github.com/hubsaude/runner/simulador/internal/processo"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibir status do Simulador HubSaude",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		s, err := processo.Consultar(ctx)
		if err != nil {
			return err
		}
		out := map[string]any{
			"registrado": s.Registrado,
			"running":    s.HealthOK,
		}
		if s.Registrado {
			out["pid"] = s.PID
			out["porta"] = s.Porta
			out["iniciadoEm"] = s.IniciadoEm
			out["pidVivo"] = s.PIDVivo
			out["versao"] = s.Versao
			if s.PIDVivo {
				out["uptimeSegundos"] = int(s.Uptime.Seconds())
			}
			if s.Info != nil {
				out["info"] = s.Info
			}
		}
		emitirJSON(cmd.OutOrStdout(), out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
