package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var iniciarCmd = &cobra.Command{
	Use:   "iniciar",
	Short: "Iniciar o Simulador HubSaude",
	Long:  "Baixa (se necessario) e inicia o simulador.jar.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando iniciar invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'iniciar' ainda nao implementado (Sprint 4)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(iniciarCmd)
	iniciarCmd.Flags().Int("porta", 0, "Porta para o simulador (0 = auto-detectar)")
	iniciarCmd.Flags().String("source", "", "URL alternativa para obter o simulador.jar")
	iniciarCmd.Flags().String("versao-simulador", "latest", "Versao especifica para baixar")
}
