package notifier

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lba-studio/n-cli/internal/config"
)

type DiscordNotifier struct {
	restyCli   *resty.Client
	configurer config.Configurer
}

type discordPayload struct {
	Content string `json:"content"`
}

var (
	ErrMissingDiscordConfig = errors.New("missing discord config")
)

func (n *DiscordNotifier) Notify(ctx context.Context, msg string) error {
	cfg, err := n.configurer.GetConfig()
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
	resp, err := n.restyCli.R().
		SetContext(ctx).
		SetBody(&payload).
		Post(url)
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to call Discord: %s", resp.String())
	}
	if err != nil {
		return err
	}
	return nil
}

func NewDiscordNotifier() Notifier {
	return &DiscordNotifier{
		restyCli: resty.New().
			SetHeader("Content-Type", "application/json").
			SetRetryCount(3).
			SetTimeout(10 * time.Second),
		configurer: config.NewConfigurer(),
	}
}
