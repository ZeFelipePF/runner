//go:build !windows

package processo

import (
	"os"
	"syscall"
)

func mandarSinal(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.SIGTERM) == nil
}
