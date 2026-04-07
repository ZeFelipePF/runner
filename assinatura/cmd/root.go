package cmd

import (
	"log/slog"

	"github.com/hubsaude/runner/assinatura/internal/logging"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
	logger  *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "assinatura",
	Short: "CLI para invocar o assinador.jar",
	Long:  "CLI multiplataforma que invoca operacoes de assinatura digital via assinador.jar.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var cleanup func()
		logger, cleanup = logging.Setup(verbose, quiet)
		cobra.OnFinalize(cleanup)
		logger.Info("assinatura CLI iniciado",
			slog.String("command", cmd.Name()),
		)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Saida detalhada (nivel debug)")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suprimir mensagens de progresso")
}
