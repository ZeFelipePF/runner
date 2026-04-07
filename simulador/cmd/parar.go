package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pararCmd = &cobra.Command{
	Use:   "parar",
	Short: "Parar o Simulador HubSaude",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando parar invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'parar' ainda nao implementado (Sprint 4)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pararCmd)
}
