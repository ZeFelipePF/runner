package assinador

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/hubsaude/runner/assinatura/internal/porta"
	"github.com/hubsaude/runner/assinatura/internal/state"
)

// Servidor reune as informacoes necessarias para falar com o assinador HTTP.
type Servidor struct {
	PID      int
	Porta    int
	Reusado  bool // true se uma instancia ja existente foi reaproveitada
	JarPath  string
	JavaPath string
}

// OpcoesStartup configura o procedimento Garantir.
type OpcoesStartup struct {
	JavaPath       string        // binario java (default: "java")
	JarPath        string        // caminho do assinador.jar
	PortaPreferida int           // porta inicial (auto-deteccao a partir dela)
	PortaForcada   int           // se > 0, exige exatamente essa porta
	TimeoutPronto  time.Duration // tempo maximo de espera pelo /health (default 30s)
	LogProgresso   io.Writer     // se nao-nil, recebe mensagens "[i] ..." e "[v] ..."
}

// Garantir verifica state.json: se ha um assinador rodando e respondendo, reusa;
// caso contrario, escolhe uma porta livre, inicia o jar e aguarda /health.
func Garantir(ctx context.Context, opc OpcoesStartup) (*Servidor, error) {
	if opc.JarPath == "" {
		return nil, errors.New("OpcoesStartup.JarPath obrigatorio")
	}
	if opc.JavaPath == "" {
		opc.JavaPath = "java"
	}
	if opc.TimeoutPronto <= 0 {
		opc.TimeoutPronto = 30 * time.Second
	}

	st, err := state.Load()
	if err != nil {
		return nil, fmt.Errorf("ao carregar state.json: %w", err)
	}
	st.CleanStale()

	if existente := tentarReusar(ctx, st, opc); existente != nil {
		progresso(opc, "Assinador em execucao na porta %d (PID %d)", existente.Porta, existente.PID)
		return existente, nil
	}

	porta, err := escolherPorta(opc)
	if err != nil {
		return nil, err
	}
	if porta != opc.PortaPreferida && opc.PortaForcada == 0 {
		progresso(opc, "Porta %d em uso. Usando porta %d.", opc.PortaPreferida, porta)
	} else {
		progresso(opc, "Iniciando assinador na porta %d...", porta)
	}

	cmd := exec.Command(opc.JavaPath, "-jar", opc.JarPath, "server", "--porta", strconv.Itoa(porta))
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ao iniciar assinador.jar: %w", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release() // desanexa: o processo continua apos o exit do CLI

	cliente := NovoClienteHTTP(porta)
	if err := cliente.AguardarPronto(ctx, opc.TimeoutPronto); err != nil {
		_ = encerrarProcesso(pid)
		return nil, fmt.Errorf("assinador nao ficou pronto: %w", err)
	}

	st.Assinador = &state.ProcessInfo{
		PID:        pid,
		Porta:      porta,
		IniciadoEm: time.Now().UTC(),
		JavaPath:   opc.JavaPath,
	}
	if err := st.Save(); err != nil {
		return nil, fmt.Errorf("ao gravar state.json: %w", err)
	}

	progresso(opc, "Assinador pronto.")
	return &Servidor{
		PID:      pid,
		Porta:    porta,
		JarPath:  opc.JarPath,
		JavaPath: opc.JavaPath,
	}, nil
}

func tentarReusar(ctx context.Context, st *state.State, opc OpcoesStartup) *Servidor {
	if st.Assinador == nil {
		return nil
	}
	info := st.Assinador
	if !state.IsProcessAlive(info.PID) {
		return nil
	}
	if opc.PortaForcada > 0 && info.Porta != opc.PortaForcada {
		return nil
	}
	cliente := NovoClienteHTTP(info.Porta)
	ctxTent, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
	defer cancel()
	if err := cliente.Health(ctxTent); err != nil {
		return nil
	}
	return &Servidor{
		PID:      info.PID,
		Porta:    info.Porta,
		Reusado:  true,
		JarPath:  opc.JarPath,
		JavaPath: info.JavaPath,
	}
}

func escolherPorta(opc OpcoesStartup) (int, error) {
	if opc.PortaForcada > 0 {
		if !porta.Disponivel(opc.PortaForcada) {
			return 0, fmt.Errorf("porta %d em uso", opc.PortaForcada)
		}
		return opc.PortaForcada, nil
	}
	inicio := opc.PortaPreferida
	if inicio <= 0 {
		inicio = 8088
	}
	return porta.LivreOuProxima(inicio)
}

func progresso(opc OpcoesStartup, formato string, args ...any) {
	if opc.LogProgresso == nil {
		return
	}
	fmt.Fprintf(opc.LogProgresso, "[i] "+formato+"\n", args...)
}

func encerrarProcesso(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
