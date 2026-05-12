// Package jdk detecta e provisiona o JDK 21 (Eclipse Temurin) localmente.
//
// Ordem de busca em Detectar():
//  1. JAVA_HOME (se versao >= VersaoMinima)
//  2. `java` no PATH
//  3. ~/.hubsaude/jdk/<distro>-<versao>/bin/java
//
// Se nenhum atender, Garantir() baixa o JDK via Adoptium.
package jdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/hubsaude/runner/assinatura/internal/state"
)

// VersaoMinima e a versao major aceita.
const VersaoMinima = 21

// JDK descreve um JDK localizado/instalado.
type JDK struct {
	JavaBin     string // caminho do executavel java
	VersaoMajor int    // ex: 21
	VersaoFull  string // ex: "21.0.3"
	Origem      string // "JAVA_HOME" | "PATH" | "hubsaude" | "adoptium"
}

// Opcoes configuram a deteccao/provisionamento.
type Opcoes struct {
	// HubsaudeDir e o diretorio base (~/.hubsaude). Se vazio, e resolvido via state.Dir().
	HubsaudeDir string
	// LogProgresso recebe mensagens de progresso. Pode ser nil.
	LogProgresso io.Writer
	// VersaoMinima sobrescreve VersaoMinima (default = pacote VersaoMinima).
	VersaoMinima int
}

// Detectar localiza um JDK ja disponivel; retorna ErrNaoEncontrado se ausente.
func Detectar(opc Opcoes) (*JDK, error) {
	min := opc.VersaoMinima
	if min == 0 {
		min = VersaoMinima
	}

	// 1. JAVA_HOME
	if javaHome := strings.TrimSpace(os.Getenv("JAVA_HOME")); javaHome != "" {
		bin := filepath.Join(javaHome, "bin", binarioJava())
		if j, ok := inspecionar(bin, min, "JAVA_HOME"); ok {
			return j, nil
		}
	}

	// 2. java no PATH
	if path, err := exec.LookPath("java"); err == nil {
		if j, ok := inspecionar(path, min, "PATH"); ok {
			return j, nil
		}
	}

	// 3. ~/.hubsaude/jdk/<*>/bin/java
	dir := opc.HubsaudeDir
	if dir == "" {
		if d, err := state.Dir(); err == nil {
			dir = d
		}
	}
	if dir != "" {
		raiz := filepath.Join(dir, "jdk")
		entradas, _ := os.ReadDir(raiz)
		for _, e := range entradas {
			if !e.IsDir() {
				continue
			}
			bin := filepath.Join(raiz, e.Name(), "bin", binarioJava())
			if j, ok := inspecionar(bin, min, "hubsaude"); ok {
				return j, nil
			}
		}
	}

	return nil, ErrNaoEncontrado
}

// Garantir retorna um JDK utilizavel — detectando local ou baixando via Adoptium.
func Garantir(ctx context.Context, opc Opcoes) (*JDK, error) {
	if j, err := Detectar(opc); err == nil {
		log(opc, "JDK detectado: %s (%s, origem=%s)", j.JavaBin, j.VersaoFull, j.Origem)
		return j, nil
	}
	log(opc, "JDK %d+ nao encontrado; baixando via Adoptium...", coalesce(opc.VersaoMinima, VersaoMinima))
	return baixarAdoptium(ctx, opc)
}

// ErrNaoEncontrado indica ausencia de JDK adequado na maquina.
var ErrNaoEncontrado = errors.New("JDK nao encontrado")

// inspecionar valida que `bin` e um java executavel >= min.
func inspecionar(bin string, min int, origem string) (*JDK, bool) {
	if bin == "" {
		return nil, false
	}
	if info, err := os.Stat(bin); err != nil || info.IsDir() {
		return nil, false
	}
	full, major, err := versaoDoBinario(bin)
	if err != nil || major < min {
		return nil, false
	}
	return &JDK{JavaBin: bin, VersaoMajor: major, VersaoFull: full, Origem: origem}, true
}

var reVersao = regexp.MustCompile(`version\s+"([^"]+)"`)

// versaoDoBinario executa `java -version` e parseia o output (stderr).
func versaoDoBinario(bin string) (string, int, error) {
	cmd := exec.Command(bin, "-version")
	saida, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0, fmt.Errorf("ao executar %s -version: %w", bin, err)
	}
	m := reVersao.FindSubmatch(saida)
	if m == nil {
		return "", 0, fmt.Errorf("formato inesperado de java -version: %q", string(saida))
	}
	full := string(m[1])
	major, err := parseMajor(full)
	if err != nil {
		return "", 0, err
	}
	return full, major, nil
}

func parseMajor(full string) (int, error) {
	// Aceita "21.0.3", "1.8.0_421" (Java 8 antigo) ou "22-ea".
	partes := strings.Split(full, ".")
	if len(partes) == 0 {
		return 0, fmt.Errorf("versao vazia")
	}
	primeiro := apenasDigitos(partes[0])
	if primeiro == "" {
		return 0, fmt.Errorf("major nao numerico em %q", full)
	}
	n, err := strconv.Atoi(primeiro)
	if err != nil {
		return 0, fmt.Errorf("major nao numerico em %q: %w", full, err)
	}
	if n == 1 && len(partes) >= 2 {
		seg := apenasDigitos(partes[1])
		if seg != "" {
			if n2, err := strconv.Atoi(seg); err == nil {
				return n2, nil
			}
		}
	}
	return n, nil
}

func apenasDigitos(s string) string {
	for i, c := range s {
		if c < '0' || c > '9' {
			return s[:i]
		}
	}
	return s
}

func binarioJava() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func log(opc Opcoes, formato string, args ...any) {
	if opc.LogProgresso == nil {
		return
	}
	fmt.Fprintf(opc.LogProgresso, "[i] "+formato+"\n", args...)
}

func coalesce(a, b int) int {
	if a > 0 {
		return a
	}
	return b
}
