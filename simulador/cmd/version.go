package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "versao",
	Short: "Exibir versao do CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("simulador %s (%s)\n", Version, GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
