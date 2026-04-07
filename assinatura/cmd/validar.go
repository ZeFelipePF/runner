package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validarCmd = &cobra.Command{
	Use:   "validar",
	Short: "Validar assinatura digital",
	Long:  "Valida uma assinatura digital utilizando o assinador.jar.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("comando validar invocado")
		fmt.Fprintln(cmd.ErrOrStderr(), "Comando 'validar' ainda nao implementado (Sprint 2+)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validarCmd)

	validarCmd.Flags().String("message-file", "", "Caminho do arquivo com conteudo original")
	validarCmd.Flags().String("signature", "", "Assinatura a validar (base64)")
	validarCmd.Flags().String("public-key", "", "Chave publica ou certificado")
	validarCmd.Flags().String("modo", "http", "Modo de invocacao: local ou http")
	validarCmd.Flags().Int("porta", 0, "Porta do servidor (0 = auto-detectar)")
}
