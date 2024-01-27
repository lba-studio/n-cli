package notifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lba-studio/n-cli/internal/config"
)

type Notifier interface {
	Notify(ctx context.Context, message string) error
}

func Notify(msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	notifierMap := map[string]Notifier{
		"system": NewSystemNotifier(),
	}

	if cfg.Discord != nil {
		notifierMap["discord"] = NewDiscordNotifier()
	}

	if cfg.Slack != nil {
		notifierMap["slack"] = NewSlackNotifier()
	}

	erroredNotifiers := make([]string, 0, len(notifierMap))
	for label, notifier := range notifierMap {
		fmt.Printf("Sending %s: ...", label)
		if err := notifier.Notify(ctx, msg); err != nil {
			fmt.Printf("ERROR (%s)\n", err.Error())
			// return err
			erroredNotifiers = append(erroredNotifiers, label)
			continue
		}
		fmt.Printf("OK\n")
	}
	if len(erroredNotifiers) > 0 {
		return fmt.Errorf("one or more notifiers failed: %s", strings.Join(erroredNotifiers, ", "))
	}
	return nil
}
