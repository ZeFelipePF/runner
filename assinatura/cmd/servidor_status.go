package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var servidorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verificar status do assinador.jar",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando servidor status invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'servidor status' ainda nao implementado (Sprint 3)")
		return nil
	},
}

func init() {
	servidorCmd.AddCommand(servidorStatusCmd)
}
