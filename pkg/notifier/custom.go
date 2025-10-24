package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lba-studio/n-cli/internal/config"
	restyutils "github.com/lba-studio/n-cli/pkg/resty_utils"
)

type CustomNotifier struct {
	restyCli   *resty.Client
	configurer config.Configurer
}

var (
	ErrCustomMissingConfig          = errors.New("missing custom config")
	ErrCustomMissingTargetUrl       = errors.New("missing targetUrl in custom config")
	ErrCustomMissingPayloadTemplate = errors.New("missing payloadTemplate in custom config")
	ErrCustomInvalidPayloadTemplate = errors.New("payloadTemplate must contain {{message}} placeholder")
	ErrCustomInvalidMethod          = errors.New("invalid HTTP method")
)

func (n *CustomNotifier) Notify(ctx context.Context, msg string) error {
	cfg, err := n.configurer.GetConfig()
	if err != nil {
		return err
	}
	if cfg.Custom == nil {
		return ErrCustomMissingConfig
	}
	if cfg.Custom.TargetUrl == "" {
		return ErrCustomMissingTargetUrl
	}
	if cfg.Custom.PayloadTemplate == "" {
		return ErrCustomMissingPayloadTemplate
	}

	// Replace {{message}} placeholder in payload template
	if !strings.Contains(cfg.Custom.PayloadTemplate, "{{message}}") {
		return ErrCustomInvalidPayloadTemplate
	}
	payloadStr := strings.Replace(cfg.Custom.PayloadTemplate, "{{message}}", msg, 1)

	// Determine HTTP method (default to POST, case-insensitive)
	method := strings.ToUpper(cfg.Custom.Method)
	if method == "" {
		method = "POST"
	}

	// Validate HTTP method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		"HEAD": true, "OPTIONS": true,
	}
	if !validMethods[method] {
		return ErrCustomInvalidMethod
	}

	// Prepare request
	req := n.restyCli.R().SetContext(ctx)

	// Apply custom headers
	for key, value := range cfg.Custom.Headers {
		req = req.SetHeader(key, value)
	}

	// Try to parse as JSON first, if it fails, send as plain text
	var payload interface{}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err == nil {
		// Valid JSON - send as JSON
		req = req.SetBody(payload)
	} else {
		// Not JSON - send as plain text
		req = req.SetBody(payloadStr)
	}

	// Make the HTTP request
	var resp *resty.Response
	switch method {
	case "GET":
		resp, err = req.Get(cfg.Custom.TargetUrl)
	case "POST":
		resp, err = req.Post(cfg.Custom.TargetUrl)
	case "PUT":
		resp, err = req.Put(cfg.Custom.TargetUrl)
	case "PATCH":
		resp, err = req.Patch(cfg.Custom.TargetUrl)
	case "DELETE":
		resp, err = req.Delete(cfg.Custom.TargetUrl)
	case "HEAD":
		resp, err = req.Head(cfg.Custom.TargetUrl)
	case "OPTIONS":
		resp, err = req.Options(cfg.Custom.TargetUrl)
	default:
		return ErrCustomInvalidMethod
	}

	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("failed to call custom webhook: %s", resp.String())
	}
	return nil
}

func NewCustomNotifier() Notifier {
	return &CustomNotifier{
		restyCli: resty.New().
			SetRetryCount(3).
			SetLogger(&restyutils.RestyLogger{}).
			SetTimeout(10 * time.Second),
		configurer: config.NewConfigurer(),
	}
}
