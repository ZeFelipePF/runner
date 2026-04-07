package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var servidorIniciarCmd = &cobra.Command{
	Use:   "iniciar",
	Short: "Iniciar o assinador.jar como servidor HTTP",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando servidor iniciar invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'servidor iniciar' ainda nao implementado (Sprint 3)")
		return nil
	},
}

func init() {
	servidorCmd.AddCommand(servidorIniciarCmd)
	servidorIniciarCmd.Flags().Int("porta", 0, "Porta do servidor (0 = auto-detectar)")
	servidorIniciarCmd.Flags().Int("timeout", 0, "Desligar apos N minutos sem uso (0 = sem timeout)")
}
