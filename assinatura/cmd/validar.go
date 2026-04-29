package cmd

import (
	"github.com/spf13/cobra"
)

var validarCmd = &cobra.Command{
	Use:   "validar",
	Short: "Validar assinatura digital",
	Long:  "Valida uma assinatura digital utilizando o assinador.jar.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executarOperacaoAssinador(cmd, "validate")
	},
}

func init() {
	rootCmd.AddCommand(validarCmd)

	validarCmd.Flags().String("payload", "", "Caminho do arquivo JSON com o payload FHIR ('-' = stdin)")
	validarCmd.Flags().String("modo", "http", "Modo de invocacao: local ou http")
	validarCmd.Flags().Int("porta", 0, "Porta do servidor HTTP (0 = usar config / auto-detectar)")
	validarCmd.Flags().String("jar", "", "Caminho do assinador.jar (default: localizar automaticamente)")
}
