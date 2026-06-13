package notifier

import (
	"context"
	"errors"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/lba-studio/n-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCustomNotifier(t *testing.T) {
	testRestyClient := resty.New()
	httpmock.ActivateNonDefault(testRestyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	defaultConfig := config.CustomConfig{
		TargetUrl:       "https://api.example.com/webhook",
		PayloadTemplate: `{"text": "{{message}}", "priority": "high"}`,
		Method:          "POST",
	}

	type testCase struct {
		name                 string
		customConfig         config.CustomConfig
		wantErr              error
		doMock               func(config.CustomConfig)
		shouldAPINotBeCalled bool
	}
	testCases := []testCase{
		{
			name:         "happy path - JSON payload template",
			customConfig: defaultConfig,
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", cfg.TargetUrl, responder)
			},
		},
		{
			name: "happy path - plain text payload template",
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: "Notification: {{message}}",
				Method:          "POST",
			},
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", cfg.TargetUrl, responder)
			},
		},
		{
			name: "happy path - different HTTP methods",
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
				Method:          "PUT",
			},
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("PUT", cfg.TargetUrl, responder)
			},
		},
		{
			name: "happy path - case insensitive method",
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
				Method:          "patch",
			},
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("PATCH", cfg.TargetUrl, responder)
			},
		},
		{
			name: "happy path - custom headers",
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
				Method:          "POST",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Custom":      "value",
				},
			},
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", cfg.TargetUrl, responder)
			},
		},
		{
			name: "happy path - default method (POST)",
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
			},
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", cfg.TargetUrl, responder)
			},
		},
		{
			name:                 "sad path - missing targetUrl",
			wantErr:              ErrCustomMissingTargetUrl,
			shouldAPINotBeCalled: true,
			customConfig: config.CustomConfig{
				PayloadTemplate: `{"message": "{{message}}"}`,
				Method:          "POST",
			},
		},
		{
			name:                 "sad path - missing payloadTemplate",
			wantErr:              ErrCustomMissingPayloadTemplate,
			shouldAPINotBeCalled: true,
			customConfig: config.CustomConfig{
				TargetUrl: "https://api.example.com/webhook",
				Method:    "POST",
			},
		},
		{
			name:                 "sad path - payloadTemplate missing {{message}} placeholder",
			wantErr:              ErrCustomInvalidPayloadTemplate,
			shouldAPINotBeCalled: true,
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"text": "no placeholder here"}`,
				Method:          "POST",
			},
		},
		{
			name:                 "sad path - invalid HTTP method",
			wantErr:              ErrCustomInvalidMethod,
			shouldAPINotBeCalled: true,
			customConfig: config.CustomConfig{
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
				Method:          "INVALID",
			},
		},
		{
			name:         "sad path - fail to call webhook",
			customConfig: defaultConfig,
			wantErr:      errors.New("failed to call custom webhook: {\"error\": \"Bad Request\"}"),
			doMock: func(cfg config.CustomConfig) {
				responder := httpmock.NewStringResponder(400, `{"error": "Bad Request"}`)
				httpmock.RegisterResponder("POST", cfg.TargetUrl, responder)
			},
		},
	}

	for _, tc := range testCases {
		httpmock.Reset()
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMock != nil {
				tc.doMock(tc.customConfig)
			}
			notifier := &CustomNotifier{
				cfg:      &tc.customConfig,
				restyCli: testRestyClient,
			}
			err := notifier.Notify(context.Background(), "my notification")
			if tc.shouldAPINotBeCalled {
				assert.Equal(t, 0, httpmock.GetTotalCallCount())
			}
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestCustomNotifierMissingConfig(t *testing.T) {
	notifier := &CustomNotifier{
		cfg: nil,
		restyCli: resty.New(),
	}
	err := notifier.Notify(context.Background(), "my notification")
	assert.Equal(t, ErrCustomMissingConfig, err)
}
