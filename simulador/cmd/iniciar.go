package cmd

import (
	"context"
	"io"
	"time"

	"github.com/hubsaude/runner/simulador/internal/download"
	"github.com/hubsaude/runner/simulador/internal/jdk"
	"github.com/hubsaude/runner/simulador/internal/processo"
	"github.com/hubsaude/runner/simulador/internal/state"
	"github.com/spf13/cobra"
)

var iniciarCmd = &cobra.Command{
	Use:   "iniciar",
	Short: "Iniciar o Simulador HubSaude",
	Long:  "Baixa (se necessario) e inicia o simulador.jar, provisionando JDK quando ausente.",
	RunE: func(cmd *cobra.Command, args []string) error {
		porta, _ := cmd.Flags().GetInt("porta")
		source, _ := cmd.Flags().GetString("source")
		versao, _ := cmd.Flags().GetString("versao-simulador")

		stderr := stderrSeNaoQuiet(cmd)

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
		defer cancel()

		// 1. Garantir JDK 21+.
		j, err := jdk.Garantir(ctx, jdk.Opcoes{LogProgresso: stderr})
		if err != nil {
			return err
		}

		// 2. Baixar/reutilizar simulador.jar.
		res, err := download.Garantir(ctx, download.Opcoes{
			Source:       source,
			Versao:       versao,
			LogProgresso: stderr,
		})
		if err != nil {
			return err
		}

		// 3. Iniciar processo.
		cfg, _ := state.LoadConfig()
		preferida := porta
		forcada := 0
		if preferida == 0 && cfg != nil {
			preferida = cfg.Simulador.PortaPadrao
		}
		if porta > 0 {
			forcada = porta
		}
		srv, err := processo.Iniciar(ctx, processo.OpcoesIniciar{
			JavaPath:       j.JavaBin,
			JarPath:        res.JarPath,
			PortaPreferida: preferida,
			PortaForcada:   forcada,
			LogProgresso:   stderr,
			Versao:         res.Metadados.Versao,
		})
		if err != nil {
			return err
		}

		status := "STARTED"
		if srv.Reusado {
			status = "REUSED"
		}
		emitirJSON(cmd.OutOrStdout(), map[string]any{
			"status":  status,
			"pid":     srv.PID,
			"porta":   srv.Porta,
			"reusado": srv.Reusado,
			"versao":  srv.Versao,
			"jar":     res.JarPath,
		})
		return nil
	},
}

// stderrSeNaoQuiet retorna cmd.ErrOrStderr() ou io.Discard quando --quiet ativo.
func stderrSeNaoQuiet(cmd *cobra.Command) io.Writer {
	if quiet {
		return io.Discard
	}
	return cmd.ErrOrStderr()
}

func init() {
	rootCmd.AddCommand(iniciarCmd)
	iniciarCmd.Flags().Int("porta", 0, "Porta para o simulador (0 = auto-detectar)")
	iniciarCmd.Flags().String("source", "", "URL alternativa para obter o simulador.jar (http://, file://)")
	iniciarCmd.Flags().String("versao-simulador", "latest", "Versao especifica para baixar")
}
