package notifier

import (
	"context"
	"errors"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/lba-studio/n-cli/internal/config"
)

type DiscordNotifier struct {
	restyCli *resty.Client
}

type discordPayload struct {
	Content string `json:"content"`
}

var (
	ErrMissingDiscordConfig = errors.New("missing discord config")
)

func (n *DiscordNotifier) Notify(ctx context.Context, msg string) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	if cfg.Discord == nil {
		return ErrMissingDiscordConfig
	}
	url := cfg.Discord.WebhookURL
	format := cfg.Discord.MessageFormat
	if format != "" {
		msg = strings.Replace(format, "{{message}}", msg, 1)
	}
	payload := discordPayload{
		Content: msg,
	}
	_, err = n.restyCli.R().
		SetBody(&payload).
		Post(url)
	if err != nil {
		return err
	}
	return nil
}

func NewDiscordNotifier() Notifier {
	return &DiscordNotifier{
		restyCli: resty.New().SetHeader("Content-Type", "application/json"),
	}
}
