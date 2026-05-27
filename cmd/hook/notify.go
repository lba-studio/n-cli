package hook

import (
	"os"

	"github.com/lba-studio/n-cli/pkg/notifier"
)

type notifyFunc func(msg string) error

var notify notifyFunc = func(msg string) error {
	return notifier.NotifyTo(msg, os.Stderr)
}
