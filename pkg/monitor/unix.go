//go:build !(linux && 386) && !windows

package monitor

import (
	"os/exec"
	"syscall"
)

// GetCPU returns the CPU time in nanoseconds
func GetCPU() (int64, error) {
	usage := new(syscall.Rusage)
	syscall.Getrusage(syscall.RUSAGE_CHILDREN, usage)
	out := usage.Utime.Nano() + usage.Stime.Nano()
	return out, nil
}

func GetMemoryFromCmd(cmd *exec.Cmd) (int64, error) {
	rusage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok {
		return 0, ErrCallNotSupported
	}
	return rusage.Maxrss, nil
}
