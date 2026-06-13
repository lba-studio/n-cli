package notifier

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lba-studio/n-cli/internal/config"
)

type Notifier interface {
	Notify(ctx context.Context, message string) error
}

func Notify(msg string) error {
	return NotifyTo(msg, os.Stdout)
}

func NotifyTo(msg string, output io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	notifierMap := map[string]Notifier{}
	if cfg.System == nil || !cfg.System.Disabled {
		notifierMap["system"] = NewSystemNotifier()
	}
	if cfg.Discord != nil {
		notifierMap["discord"] = NewDiscordNotifier()
	}
	if cfg.Slack != nil {
		notifierMap["slack"] = NewSlackNotifier()
	}
	for _, entry := range customNotifierEntries(cfg) {
		notifierMap[entry.label] = NewCustomNotifierFromConfig(entry.cfg)
	}
	labels := make([]string, 0, len(notifierMap))
	for label := range notifierMap {
		labels = append(labels, label)
	}

	fmt.Fprintf(output, "Sending notification to %d channels: %s\n", len(notifierMap), strings.Join(labels, ", "))

	erroredNotifiers := make([]string, 0, len(notifierMap))
	wg := sync.WaitGroup{}
	erroredNotifiersChan := make(chan string, len(notifierMap))
	for label, notifier := range notifierMap {
		wg.Add(1)
		go func(label string, notifier Notifier) {
			defer wg.Done()
			logPrefix := fmt.Sprintf("Sent notification to %s", label)
			if err := notifier.Notify(ctx, msg); err != nil {
				fmt.Fprintf(output, "%s...ERROR (%s)\n", logPrefix, err.Error())
				erroredNotifiersChan <- label
				return
			}
			fmt.Fprintf(output, "%s...OK\n", logPrefix)
		}(label, notifier)
	}
	go func() {
		wg.Wait()
		close(erroredNotifiersChan)
	}()

	for {
		erroredNotifier, ok := <-erroredNotifiersChan
		if !ok {
			break
		}
		erroredNotifiers = append(erroredNotifiers, erroredNotifier)
	}

	if len(erroredNotifiers) > 0 {
		return fmt.Errorf("one or more notifiers failed: %s", strings.Join(erroredNotifiers, ", "))
	}
	return nil
}

type customNotifierEntry struct {
	label string
	cfg   config.CustomConfig
}

func customNotifierEntries(cfg config.Config) []customNotifierEntry {
	var customConfigs []config.CustomConfig
	if cfg.Custom != nil {
		customConfigs = append(customConfigs, *cfg.Custom)
	}
	customConfigs = append(customConfigs, cfg.Customs...)

	entries := make([]customNotifierEntry, 0, len(customConfigs))
	for i, cc := range customConfigs {
		label := cc.Name
		if label == "" {
			label = fmt.Sprintf("custom[%d]", i)
		}
		entries = append(entries, customNotifierEntry{
			label: label,
			cfg:   cc,
		})
	}
	return entries
}
