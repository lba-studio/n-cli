package notifier

import (
	"context"
	"fmt"
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

	for label, notifier := range notifierMap {
		fmt.Printf("Sending %s: ...", label)
		if err := notifier.Notify(ctx, msg); err != nil {
			fmt.Printf("ERROR (%s)\n", err.Error())
			return err
		}
		fmt.Printf("OK\n")
	}
	return nil
}
