//go:build !windows

package monitor

func IsWindows() bool {
	return false
}
