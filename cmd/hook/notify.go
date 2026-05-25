package hook

import "github.com/lba-studio/n-cli/pkg/notifier"

type notifyFunc func(msg string) error

var notify notifyFunc = notifier.Notify
