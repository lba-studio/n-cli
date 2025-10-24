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
	defaultConfig := config.Config{
		Custom: &config.CustomConfig{
			TargetUrl:       "https://api.example.com/webhook",
			PayloadTemplate: `{"text": "{{message}}", "priority": "high"}`,
			Method:          "POST",
		},
	}
	var mockConfigurer *config.MockConfigurer

	type testCase struct {
		name                 string
		wantErr              error
		doMock               func()
		shouldAPINotBeCalled bool
	}
	testCases := []testCase{
		{
			name: "happy path - JSON payload template",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(defaultConfig, nil)
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", defaultConfig.Custom.TargetUrl, responder)
			},
		},
		{
			name: "happy path - plain text payload template",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: "Notification: {{message}}",
						Method:          "POST",
					}}, nil)
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", defaultConfig.Custom.TargetUrl, responder)
			},
		},
		{
			name: "happy path - different HTTP methods",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: `{"message": "{{message}}"}`,
						Method:          "PUT",
					}}, nil)
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("PUT", defaultConfig.Custom.TargetUrl, responder)
			},
		},
		{
			name: "happy path - case insensitive method",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: `{"message": "{{message}}"}`,
						Method:          "patch",
					}}, nil)
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("PATCH", defaultConfig.Custom.TargetUrl, responder)
			},
		},
		{
			name: "happy path - custom headers",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: `{"message": "{{message}}"}`,
						Method:          "POST",
						Headers: map[string]string{
							"Authorization": "Bearer token123",
							"X-Custom":      "value",
						},
					}}, nil)
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", defaultConfig.Custom.TargetUrl, responder)
			},
		},
		{
			name: "happy path - default method (POST)",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: `{"message": "{{message}}"}`,
						// Method not specified, should default to POST
					}}, nil)
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", defaultConfig.Custom.TargetUrl, responder)
			},
		},
		{
			name: "do not run when custom has no config",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{}, nil)
			},
			wantErr:              ErrCustomMissingConfig,
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - cannot get config",
			wantErr: errors.New("config error"),
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{}, errors.New("config error"))
			},
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - missing targetUrl",
			wantErr: ErrCustomMissingTargetUrl,
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						PayloadTemplate: `{"message": "{{message}}"}`,
						Method:          "POST",
					}}, nil)
			},
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - missing payloadTemplate",
			wantErr: ErrCustomMissingPayloadTemplate,
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl: "https://api.example.com/webhook",
						Method:    "POST",
					}}, nil)
			},
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - payloadTemplate missing {{message}} placeholder",
			wantErr: ErrCustomInvalidPayloadTemplate,
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: `{"text": "no placeholder here"}`,
						Method:          "POST",
					}}, nil)
			},
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - invalid HTTP method",
			wantErr: ErrCustomInvalidMethod,
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Custom: &config.CustomConfig{
						TargetUrl:       "https://api.example.com/webhook",
						PayloadTemplate: `{"message": "{{message}}"}`,
						Method:          "INVALID",
					}}, nil)
			},
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - fail to call webhook",
			wantErr: errors.New("failed to call custom webhook: {\"error\": \"Bad Request\"}"),
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(defaultConfig, nil)
				responder := httpmock.NewStringResponder(400, `{"error": "Bad Request"}`)
				httpmock.RegisterResponder("POST", defaultConfig.Custom.TargetUrl, responder)
			},
		},
	}

	for _, tc := range testCases {
		httpmock.Reset()
		mockConfigurer = config.NewMockConfigurer(t)
		t.Run(tc.name, func(t *testing.T) {
			tc.doMock()
			notifier := &CustomNotifier{
				restyCli:   testRestyClient,
				configurer: mockConfigurer,
			}
			err := notifier.Notify(context.Background(), "my notification")
			if tc.shouldAPINotBeCalled {
				assert.Equal(t, 0, httpmock.GetTotalCallCount())
			}
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
