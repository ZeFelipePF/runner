package cmd

import (
	"log/slog"

	"github.com/hubsaude/runner/simulador/internal/logging"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
	logger  *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "simulador",
	Short: "CLI para gerenciar o Simulador HubSaude",
	Long:  "CLI multiplataforma que gerencia o ciclo de vida do simulador.jar do HubSaude.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var cleanup func()
		logger, cleanup = logging.Setup(verbose, quiet)
		cobra.OnFinalize(cleanup)
		logger.Info("simulador CLI iniciado",
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
