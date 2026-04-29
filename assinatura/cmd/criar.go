package cmd

import (
	"github.com/spf13/cobra"
)

var criarCmd = &cobra.Command{
	Use:   "criar",
	Short: "Criar assinatura digital",
	Long:  "Cria uma assinatura digital utilizando o assinador.jar.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executarOperacaoAssinador(cmd, "sign")
	},
}

func init() {
	rootCmd.AddCommand(criarCmd)

	criarCmd.Flags().String("payload", "", "Caminho do arquivo JSON com o payload FHIR ('-' = stdin)")
	criarCmd.Flags().String("modo", "http", "Modo de invocacao: local ou http")
	criarCmd.Flags().Int("porta", 0, "Porta do servidor HTTP (0 = usar config / auto-detectar)")
	criarCmd.Flags().String("jar", "", "Caminho do assinador.jar (default: localizar automaticamente)")
}
