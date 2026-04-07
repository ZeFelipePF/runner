package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibir status do Simulador HubSaude",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando status invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'status' ainda nao implementado (Sprint 4)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
