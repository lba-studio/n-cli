package monitor

import "syscall"

// GetCPU returns the CPU time in nanoseconds
func GetCPU() int64 {
	usage := new(syscall.Rusage)
	syscall.Getrusage(syscall.RUSAGE_CHILDREN, usage)
	return usage.Utime.Nano() + usage.Stime.Nano()
}
