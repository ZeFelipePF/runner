// Package porta auto-detecta portas TCP livres no localhost.
package porta

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// JanelaProcura define o range a partir da porta inicial: [inicio, inicio+JanelaProcura].
const JanelaProcura = 20

// LivreOuProxima retorna a primeira porta TCP disponivel a partir de inicio
// ate inicio+JanelaProcura (inclusive). Erro se todas estiverem ocupadas.
func LivreOuProxima(inicio int) (int, error) {
	if inicio <= 0 || inicio > 65535 {
		return 0, fmt.Errorf("porta inicial invalida: %d", inicio)
	}
	for p := inicio; p <= inicio+JanelaProcura && p <= 65535; p++ {
		if Disponivel(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("nenhuma porta livre no range %d-%d", inicio, inicio+JanelaProcura)
}

// Disponivel verifica se e possivel ouvir em localhost:porta.
func Disponivel(porta int) bool {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(porta))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

// EmUso e o oposto de Disponivel.
func EmUso(porta int) bool {
	return !Disponivel(porta)
}

// AguardarEmUso bloqueia ate que a porta esteja escutando ou timeout expire.
// Retorna true se a porta foi detectada em uso.
func AguardarEmUso(porta int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(porta))
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
