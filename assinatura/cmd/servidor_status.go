package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hubsaude/runner/assinatura/internal/assinador"
	"github.com/hubsaude/runner/assinatura/internal/state"
	"github.com/spf13/cobra"
)

var servidorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verificar status do assinador.jar",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := state.Load()
		if err != nil {
			return err
		}

		out := map[string]any{"running": false}
		if st.Assinador != nil {
			out["pid"] = st.Assinador.PID
			out["porta"] = st.Assinador.Porta
			out["iniciadoEm"] = st.Assinador.IniciadoEm

			vivo := state.IsProcessAlive(st.Assinador.PID)
			out["pidVivo"] = vivo

			if vivo {
				ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
				defer cancel()
				if err := assinador.NovoClienteHTTP(st.Assinador.Porta).Health(ctx); err == nil {
					out["running"] = true
				}
			}
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	},
}

func init() {
	servidorCmd.AddCommand(servidorStatusCmd)
}
