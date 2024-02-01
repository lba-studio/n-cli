//go:build windows

package monitor

import (
	"os/exec"
)

// GetCPU returns the CPU time in nanoseconds
func GetCPU() (int64, error) {
	// not supported (for now?)
	return 0, nil
}

func GetMemoryFromCmd(cmd *exec.Cmd) (int64, error) {
	// not supported (for now?)
	return 0, nil
}
