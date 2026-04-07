package cmd

import "github.com/spf13/cobra"

var servidorCmd = &cobra.Command{
	Use:   "servidor",
	Short: "Gerenciar o assinador.jar no modo servidor",
	Long:  "Subcomandos para iniciar, parar e consultar o status do assinador.jar.",
}

func init() {
	rootCmd.AddCommand(servidorCmd)
}
