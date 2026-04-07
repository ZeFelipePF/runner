package main

import (
	"os"

	"github.com/hubsaude/runner/assinatura/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
