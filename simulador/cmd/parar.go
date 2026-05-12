package cmd

import (
	"context"
	"time"

	"github.com/hubsaude/runner/simulador/internal/processo"
	"github.com/spf13/cobra"
)

var pararCmd = &cobra.Command{
	Use:   "parar",
	Short: "Parar o Simulador HubSaude",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()
		r, err := processo.Parar(ctx)
		if err != nil {
			return err
		}
		status := "STOPPED"
		if r.Metodo == "nao_estava_rodando" {
			status = "NOT_RUNNING"
		}
		emitirJSON(cmd.OutOrStdout(), map[string]any{
			"status": status,
			"pid":    r.PID,
			"porta":  r.Porta,
			"metodo": r.Metodo,
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pararCmd)
}
