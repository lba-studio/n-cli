package monitor

import "errors"

var (
	ErrCallNotSupported = errors.New("this syscall / operation is unsupported")
)
