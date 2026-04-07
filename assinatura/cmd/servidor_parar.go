package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var servidorPararCmd = &cobra.Command{
	Use:   "parar",
	Short: "Parar o assinador.jar",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando servidor parar invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'servidor parar' ainda nao implementado (Sprint 3)")
		return nil
	},
}

func init() {
	servidorCmd.AddCommand(servidorPararCmd)
	servidorPararCmd.Flags().Int("porta", 0, "Porta do servidor a parar")
}
