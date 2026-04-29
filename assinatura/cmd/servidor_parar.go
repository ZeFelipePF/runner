package cmd

import (
	"context"
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
			fmt.Fprintln(cmd.ErrOrStderr(), "[i] nenhum assinador registrado em state.json")
			return nil
		}

		porta := st.Assinador.Porta
		pid := st.Assinador.PID

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		cliente := assinador.NovoClienteHTTP(porta)
		shutdownErr := cliente.Shutdown(ctx)
		if shutdownErr != nil && state.IsProcessAlive(pid) {
			if p, e := os.FindProcess(pid); e == nil {
				_ = p.Kill()
			}
		}

		st.Assinador = nil
		if err := st.Save(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), `{"status":"STOPPED"}`)
		return nil
	},
}

func init() {
	servidorCmd.AddCommand(servidorPararCmd)
	servidorPararCmd.Flags().Int("porta", 0, "Porta do servidor a parar (informativa)")
}
