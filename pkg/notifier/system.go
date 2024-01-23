package notifier

import (
	"context"

	"github.com/gen2brain/beeep"
)

type SystemNotifier struct{}

func (n *SystemNotifier) Notify(ctx context.Context, msg string) error {
	return beeep.Notify("N: New Notification", msg, "")
}

func NewSystemNotifier() Notifier {
	return &SystemNotifier{}
}
