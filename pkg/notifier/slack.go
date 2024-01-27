package notifier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lba-studio/n-cli/internal/config"
	"github.com/lba-studio/n-cli/pkg/notifier/utils"
	"github.com/lba-studio/n-cli/pkg/notifier/webhook"
	restyutils "github.com/lba-studio/n-cli/pkg/resty_utils"
)

type SlackNotifier struct {
	restyCli   *resty.Client
	configurer config.Configurer
}

type slackPayload struct {
	Message string `json:"message"`
}

type slackResponse struct {
	OK bool `json:"ok"`
}

var (
	ErrSlackMissingConfig = errors.New("missing slack config")
)

func (n *SlackNotifier) Notify(ctx context.Context, msg string) error {
	cfg, err := n.configurer.GetConfig()
	if err != nil {
		return err
	}
	if cfg.Slack == nil {
		return ErrSlackMissingConfig
	}
	url := cfg.Slack.WebhookURL
	if url == "" {
		return webhook.ErrWebhookMissingWebhookURL
	}
	format := cfg.Slack.MessageFormat
	msg, err = utils.GetMessageFromFormat(format, msg)
	if err != nil {
		return err
	}
	payload := slackPayload{
		Message: msg,
	}
	resp, err := n.restyCli.R().
		SetContext(ctx).
		SetBody(&payload).
		SetResult(slackResponse{}).
		Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to call Slack: %s", resp.String())
	}
	responseBody := *resp.Result().(*slackResponse)
	if !responseBody.OK {
		return fmt.Errorf("responseBody not ok: responseBody=%+v resp.String=%s", responseBody, resp.String())
	}
	return nil
}

func NewSlackNotifier() Notifier {
	return &SlackNotifier{
		restyCli: resty.New().
			SetHeader("Content-Type", "application/json").
			SetRetryCount(3).
			SetLogger(&restyutils.RestyLogger{}).
			SetTimeout(10 * time.Second),
		configurer: config.NewConfigurer(),
	}
}
