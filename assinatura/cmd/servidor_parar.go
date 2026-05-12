package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hubsaude/runner/assinatura/internal/assinador"
	"github.com/hubsaude/runner/assinatura/internal/state"
	"github.com/spf13/cobra"
)

var servidorPararCmd = &cobra.Command{
	Use:   "parar",
	Short: "Parar o assinador.jar",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := state.Load()
		if err != nil {
			return err
		}
		if st.Assinador == nil {
			emitirJSON(cmd.OutOrStdout(), map[string]any{
				"status":  "NOT_RUNNING",
				"message": "nenhum assinador registrado em state.json",
			})
			return nil
		}

		porta := st.Assinador.Porta
		pid := st.Assinador.PID

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		metodo := "http_shutdown"
		cliente := assinador.NovoClienteHTTP(porta)
		shutdownErr := cliente.Shutdown(ctx)
		if shutdownErr != nil {
			metodo = "sigkill"
			if state.IsProcessAlive(pid) {
				if p, e := os.FindProcess(pid); e == nil {
					_ = p.Kill()
				}
			} else {
				metodo = "stale_pid"
			}
		}

		st.Assinador = nil
		if err := st.Save(); err != nil {
			return fmt.Errorf("ao limpar state.json: %w", err)
		}

		emitirJSON(cmd.OutOrStdout(), map[string]any{
			"status":  "STOPPED",
			"pid":     pid,
			"porta":   porta,
			"metodo":  metodo,
		})
		return nil
	},
}

func emitirJSON(w interface{ Write([]byte) (int, error) }, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func init() {
	servidorCmd.AddCommand(servidorPararCmd)
	servidorPararCmd.Flags().Int("porta", 0, "Porta do servidor a parar (informativa)")
}
