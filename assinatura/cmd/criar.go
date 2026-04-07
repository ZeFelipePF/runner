package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var criarCmd = &cobra.Command{
	Use:   "criar",
	Short: "Criar assinatura digital",
	Long:  "Cria uma assinatura digital utilizando o assinador.jar.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando criar invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'criar' ainda nao implementado (Sprint 2+)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(criarCmd)

	criarCmd.Flags().String("message-file", "", "Caminho do arquivo com conteudo a assinar")
	criarCmd.Flags().String("private-key", "", "Caminho do arquivo de chave privada")
	criarCmd.Flags().String("algoritmo", "SHA256withRSA", "Algoritmo de assinatura")
	criarCmd.Flags().String("modo", "http", "Modo de invocacao: local ou http")
	criarCmd.Flags().Int("porta", 0, "Porta do servidor (0 = auto-detectar)")
}
