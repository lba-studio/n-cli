package restyutils

import "github.com/go-resty/resty/v2"

type RestyLogger struct{}

// Debugf implements resty.Logger.
func (*RestyLogger) Debugf(format string, v ...interface{}) {
}

// Errorf implements resty.Logger.
func (*RestyLogger) Errorf(format string, v ...interface{}) {
}

// Warnf implements resty.Logger.
func (*RestyLogger) Warnf(format string, v ...interface{}) {
}

var _ resty.Logger = &RestyLogger{}
