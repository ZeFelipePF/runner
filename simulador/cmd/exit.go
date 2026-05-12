package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// emitirJSON serializa v como JSON pretty-printed no writer dado.
func emitirJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// EmitirErroJSON escreve {"error": "...", "message": "..."} em stderr.
func EmitirErroJSON(w io.Writer, err error) {
	if err == nil {
		return
	}
	emitirJSON(w, map[string]any{
		"error":   "CLI_ERROR",
		"message": err.Error(),
	})
}

// ExecuteOrExit roda a CLI e propaga exit codes estruturados.
// Exit 0 sucesso, 1 erro generico, 2 argumentos invalidos (cobra cuida).
func ExecuteOrExit() {
	if err := Execute(); err != nil {
		EmitirErroJSON(os.Stderr, err)
		fmt.Fprintln(os.Stderr) // garantir newline final no shell.
		os.Exit(1)
	}
}
