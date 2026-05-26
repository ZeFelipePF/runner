// Package processo gerencia o ciclo de vida do simulador.jar:
// startup com detecao de instancia existente, parada via HTTP + fallback SIGTERM
// e consulta de status.
package processo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/hubsaude/runner/simulador/internal/porta"
	"github.com/hubsaude/runner/simulador/internal/state"
)

// Servidor representa um simulador.jar em execucao.
type Servidor struct {
	PID      int
	Porta    int
	Reusado  bool
	JarPath  string
	JavaPath string
	Versao   string
}

// OpcoesIniciar configuram um startup do simulador.
type OpcoesIniciar struct {
	JavaPath       string
	JarPath        string
	PortaPreferida int
	PortaForcada   int
	TimeoutPronto  time.Duration
	LogProgresso   io.Writer
	Versao         string
}

// Iniciar reusa instancia existente (state.json + PID + /actuator/health) ou
// dispara um novo `java -jar simulador.jar --server.port <porta>`.
func Iniciar(ctx context.Context, opc OpcoesIniciar) (*Servidor, error) {
	if opc.JarPath == "" {
		return nil, errors.New("OpcoesIniciar.JarPath obrigatorio")
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
		log(opc, "Simulador em execucao na porta %d (PID %d)", existente.Porta, existente.PID)
		return existente, nil
	}

	p, err := escolherPorta(opc)
	if err != nil {
		return nil, err
	}
	if p != opc.PortaPreferida && opc.PortaForcada == 0 {
		log(opc, "Porta %d em uso. Usando porta %d.", opc.PortaPreferida, p)
	} else {
		log(opc, "Iniciando simulador na porta %d...", p)
	}

	cmd := exec.Command(opc.JavaPath, "-jar", opc.JarPath, "--server.port="+strconv.Itoa(p))
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ao iniciar simulador.jar: %w", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()

	if err := aguardarPorta(ctx, p, opc.TimeoutPronto); err != nil {
		_ = mata(pid)
		return nil, fmt.Errorf("simulador nao ficou pronto: %w", err)
	}

	st.Simulador = &state.ProcessInfo{
		PID:        pid,
		Porta:      p,
		IniciadoEm: time.Now().UTC(),
		JavaPath:   opc.JavaPath,
		Versao:     opc.Versao,
	}
	if err := st.Save(); err != nil {
		return nil, fmt.Errorf("ao gravar state.json: %w", err)
	}

	log(opc, "Simulador pronto.")
	return &Servidor{
		PID:      pid,
		Porta:    p,
		JarPath:  opc.JarPath,
		JavaPath: opc.JavaPath,
		Versao:   opc.Versao,
	}, nil
}

func tentarReusar(ctx context.Context, st *state.State, opc OpcoesIniciar) *Servidor {
	if st.Simulador == nil {
		return nil
	}
	info := st.Simulador
	if !state.IsProcessAlive(info.PID) {
		return nil
	}
	if opc.PortaForcada > 0 && info.Porta != opc.PortaForcada {
		return nil
	}
	ctxTent, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
	defer cancel()
	if !portaResponde(ctxTent, info.Porta) {
		return nil
	}
	return &Servidor{
		PID:      info.PID,
		Porta:    info.Porta,
		Reusado:  true,
		JavaPath: info.JavaPath,
		Versao:   info.Versao,
	}
}

func escolherPorta(opc OpcoesIniciar) (int, error) {
	if opc.PortaForcada > 0 {
		if !porta.Disponivel(opc.PortaForcada) {
			return 0, fmt.Errorf("porta %d em uso", opc.PortaForcada)
		}
		return opc.PortaForcada, nil
	}
	inicio := opc.PortaPreferida
	if inicio <= 0 {
		inicio = state.PortaPadraoSimulador
	}
	return porta.LivreOuProxima(inicio)
}

// aguardarPorta bloqueia ate que a porta esteja escutando ou timeout estoure.
func aguardarPorta(ctx context.Context, p int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !porta.Disponivel(p) {
			// Esta em uso => alguem (provavelmente o processo recem-iniciado) escuta.
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("porta %d nao foi escutada em %s", p, timeout)
}

func portaResponde(ctx context.Context, p int) bool {
	// Tenta /actuator/health (Spring) — caminho HubSaude. Se nao responder,
	// considera "vivo" se a porta esta em uso.
	urls := []string{
		"http://127.0.0.1:" + strconv.Itoa(p) + "/actuator/health",
		"http://127.0.0.1:" + strconv.Itoa(p) + "/health",
	}
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode/100 == 2 {
			return true
		}
	}
	return porta.EmUso(p)
}

// ResultadoParada descreve como o simulador foi encerrado.
type ResultadoParada struct {
	Metodo string // "http_shutdown" | "sigterm" | "sigkill" | "stale_pid" | "nao_estava_rodando"
	PID    int
	Porta  int
}

// Parar encerra o simulador registrado em state.json.
func Parar(ctx context.Context) (*ResultadoParada, error) {
	st, err := state.Load()
	if err != nil {
		return nil, err
	}
	if st.Simulador == nil {
		return &ResultadoParada{Metodo: "nao_estava_rodando"}, nil
	}
	pid := st.Simulador.PID
	p := st.Simulador.Porta
	res := &ResultadoParada{PID: pid, Porta: p}

	if !state.IsProcessAlive(pid) {
		res.Metodo = "stale_pid"
	} else if tentarHTTPShutdown(ctx, p) && esperarMorte(pid, 3*time.Second) {
		res.Metodo = "http_shutdown"
	} else if mandarSinal(pid) && esperarMorte(pid, 5*time.Second) {
		res.Metodo = "sigterm"
	} else {
		_ = mata(pid)
		res.Metodo = "sigkill"
	}

	st.Simulador = nil
	if err := st.Save(); err != nil {
		return res, fmt.Errorf("ao limpar state.json: %w", err)
	}
	return res, nil
}

func tentarHTTPShutdown(ctx context.Context, p int) bool {
	urls := []string{
		"http://127.0.0.1:" + strconv.Itoa(p) + "/actuator/shutdown",
		"http://127.0.0.1:" + strconv.Itoa(p) + "/shutdown",
	}
	client := &http.Client{Timeout: 2 * time.Second}
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode/100 == 2 {
			return true
		}
	}
	return false
}

func esperarMorte(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !state.IsProcessAlive(pid) {
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return !state.IsProcessAlive(pid)
}

// Status descreve o estado atual do simulador segundo state.json.
type Status struct {
	Registrado bool
	PID        int
	Porta      int
	IniciadoEm time.Time
	PIDVivo    bool
	HealthOK   bool
	Versao     string
	Uptime     time.Duration
	// Info e o payload de GET /api/info do simulador.jar quando disponivel
	// (criterio US-03 da especificacao). Nulo se o endpoint nao responder.
	Info map[string]any
}

// Consultar le state.json e enriquece com checagem de PID/health.
func Consultar(ctx context.Context) (*Status, error) {
	st, err := state.Load()
	if err != nil {
		return nil, err
	}
	out := &Status{}
	if st.Simulador == nil {
		return out, nil
	}
	out.Registrado = true
	out.PID = st.Simulador.PID
	out.Porta = st.Simulador.Porta
	out.IniciadoEm = st.Simulador.IniciadoEm
	out.Versao = st.Simulador.Versao
	out.PIDVivo = state.IsProcessAlive(out.PID)
	if out.PIDVivo {
		out.HealthOK = portaResponde(ctx, out.Porta)
		out.Uptime = time.Since(out.IniciadoEm)
		out.Info = consultarApiInfo(ctx, out.Porta)
	}
	return out, nil
}

// consultarApiInfo busca GET /api/info (criterio US-03 da especificacao).
// Retorna nil se o endpoint nao existir ou nao responder JSON valido.
func consultarApiInfo(ctx context.Context, p int) map[string]any {
	url := "http://127.0.0.1:" + strconv.Itoa(p) + "/api/info"
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(c, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil
	}
	var info map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&info); err != nil {
		return nil
	}
	return info
}

func mata(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

func log(opc OpcoesIniciar, formato string, args ...any) {
	if opc.LogProgresso == nil {
		return
	}
	fmt.Fprintf(opc.LogProgresso, "[i] "+formato+"\n", args...)
}
