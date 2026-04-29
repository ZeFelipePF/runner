package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/hubsaude/runner/assinatura/internal/assinador"
	"github.com/hubsaude/runner/assinatura/internal/state"
	"github.com/spf13/cobra"
)

var servidorIniciarCmd = &cobra.Command{
	Use:   "iniciar",
	Short: "Iniciar o assinador.jar como servidor HTTP",
	RunE: func(cmd *cobra.Command, args []string) error {
		porta, _ := cmd.Flags().GetInt("porta")
		jar, _ := cmd.Flags().GetString("jar")

		jarPath := jar
		if jarPath == "" {
			p, err := assinador.LocalizarJar()
			if err != nil {
				return err
			}
			jarPath = p
		}

		cfg, _ := state.LoadConfig()
		preferida := porta
		forcada := 0
		if preferida == 0 && cfg != nil {
			preferida = cfg.Assinador.PortaPadrao
		}
		if porta > 0 {
			forcada = porta
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
		defer cancel()

		srv, err := assinador.Garantir(ctx, assinador.OpcoesStartup{
			JarPath:        jarPath,
			PortaPreferida: preferida,
			PortaForcada:   forcada,
			LogProgresso:   cmd.ErrOrStderr(),
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "{\"status\":\"OK\",\"porta\":%d,\"pid\":%d,\"reusado\":%t}\n",
			srv.Porta, srv.PID, srv.Reusado)
		return nil
	},
}

func init() {
	servidorCmd.AddCommand(servidorIniciarCmd)
	servidorIniciarCmd.Flags().Int("porta", 0, "Porta do servidor (0 = config / auto-detectar)")
	servidorIniciarCmd.Flags().String("jar", "", "Caminho do assinador.jar")
	servidorIniciarCmd.Flags().Int("timeout", 0, "Desligar apos N minutos sem uso (0 = sem timeout) — nao implementado")
}
