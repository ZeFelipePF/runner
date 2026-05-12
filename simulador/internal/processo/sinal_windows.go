//go:build windows

package processo

// No Windows nao ha SIGTERM cooperativo: usar Kill direto ja e suficiente.
func mandarSinal(pid int) bool {
	return false
}
