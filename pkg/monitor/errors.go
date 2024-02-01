package monitor

import "errors"

var (
	ErrCallNotSupported = errors.New("this syscall / operation is unsupported")
	ErrIsWindows        = errors.New("not supported on Windows-based machines")
)
