package notifier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lba-studio/n-cli/internal/config"
	"github.com/lba-studio/n-cli/pkg/notifier/utils"
	restyutils "github.com/lba-studio/n-cli/pkg/resty_utils"
)

type DiscordNotifier struct {
	restyCli   *resty.Client
	configurer config.Configurer
}

type discordPayload struct {
	Content string `json:"content"`
}

var (
	ErrDiscordMissingConfig            = errors.New("missing discord config")
	ErrDiscordFormatMissingPlaceholder = errors.New("{{message}} placeholder is missing from messageFormat")
)

func (n *DiscordNotifier) Notify(ctx context.Context, msg string) error {
	cfg, err := n.configurer.GetConfig()
	if err != nil {
		return err
	}
	if cfg.Discord == nil {
		return ErrDiscordMissingConfig
	}
	url := cfg.Discord.WebhookURL
	format := cfg.Discord.MessageFormat
	msg, err = utils.GetMessageFromFormat(format, msg)
	if err != nil {
		return err
	}
	payload := discordPayload{
		Content: msg,
	}
	resp, err := n.restyCli.R().
		SetContext(ctx).
		SetBody(&payload).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to call Discord: %s", resp.String())
	}
	return nil
}

func NewDiscordNotifier() Notifier {
	return &DiscordNotifier{
		restyCli: resty.New().
			SetHeader("Content-Type", "application/json").
			SetRetryCount(3).
			SetLogger(&restyutils.RestyLogger{}).
			SetTimeout(10 * time.Second),
		configurer: config.NewConfigurer(),
	}
}
