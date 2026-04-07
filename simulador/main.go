package main

import (
	"os"

	"github.com/hubsaude/runner/simulador/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
