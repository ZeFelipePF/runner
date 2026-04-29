package porta

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisponivel_PortaLivre(t *testing.T) {
	p := portaLivre(t)
	assert.True(t, Disponivel(p))
}

func TestDisponivel_PortaOcupada(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	porta := l.Addr().(*net.TCPAddr).Port
	assert.False(t, Disponivel(porta))
	assert.True(t, EmUso(porta))
}

func TestLivreOuProxima_RetornaInicioSeLivre(t *testing.T) {
	inicio := portaLivre(t)
	p, err := LivreOuProxima(inicio)
	require.NoError(t, err)
	assert.Equal(t, inicio, p)
}

func TestLivreOuProxima_AvancaSeOcupada(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	inicio := l.Addr().(*net.TCPAddr).Port

	p, err := LivreOuProxima(inicio)
	require.NoError(t, err)
	assert.Greater(t, p, inicio)
	assert.LessOrEqual(t, p, inicio+JanelaProcura)
}

func TestLivreOuProxima_PortaInvalida(t *testing.T) {
	_, err := LivreOuProxima(0)
	assert.Error(t, err)
	_, err = LivreOuProxima(70000)
	assert.Error(t, err)
}

// portaLivre retorna uma porta TCP livre no momento da chamada.
func portaLivre(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	porta := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return porta
}
