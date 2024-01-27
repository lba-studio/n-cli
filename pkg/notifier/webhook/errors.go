package webhook

import "errors"

var (
	ErrWebhookMissingWebhookURL = errors.New("missing webhookUrl")
)
