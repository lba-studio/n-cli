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

func TestSlackNotifier(t *testing.T) {
	testRestyClient := resty.New()
	httpmock.ActivateNonDefault(testRestyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	defaultConfig := config.Config{
		Slack: &config.SlackConfig{
			WebhookURL: "https://blah.com",
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
			name: "happy path",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(defaultConfig, nil)
				// responder := httpmock.NewStringResponder(200, `{"ok": true}`).HeaderAdd(map[string][]string{
				// 	"Content-Type": []string{"application/json"},
				// })
				responder := httpmock.NewJsonResponderOrPanic(200, map[string]any{"ok": true})
				httpmock.RegisterResponder("POST", defaultConfig.Slack.WebhookURL, responder)
			},
		},
		{
			name: "happy path - with messageFormat",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Slack: &config.SlackConfig{
						WebhookURL:    "https://blah.com",
						MessageFormat: "my message is {{message}}",
					}}, nil)
				responder := httpmock.NewJsonResponderOrPanic(200, map[string]any{"ok": true})
				httpmock.RegisterResponder("POST", defaultConfig.Slack.WebhookURL, responder)
			},
		},
		{
			name: "do not run when slack has no config",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{}, nil)
			},
			wantErr:              ErrSlackMissingConfig,
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - cannot get config",
			wantErr: errors.New("oop"),
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{}, errors.New("oop"))
			},
			shouldAPINotBeCalled: true,
		},
		{
			name:    "sad path - fail to call slack",
			wantErr: errors.New("failed to call Slack: {\"message\": \"Yo this is broken\"}"),
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(defaultConfig, nil)
				responder := httpmock.NewStringResponder(400, `{"message": "Yo this is broken"}`)
				httpmock.RegisterResponder("POST", defaultConfig.Slack.WebhookURL, responder)
			},
		},
		{
			name:    "sad path - slack config has messageFormat but it's missing the message placeholder",
			wantErr: errors.New("{{message}} placeholder is missing from messageFormat"),
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{
					Slack: &config.SlackConfig{
						WebhookURL:    "https://blah.com",
						MessageFormat: "oop this has no placeholder",
					}}, nil)
			},
			shouldAPINotBeCalled: true,
		},
	}

	for _, tc := range testCases {
		httpmock.Reset()
		mockConfigurer = config.NewMockConfigurer(t)
		t.Run(tc.name, func(t *testing.T) {
			tc.doMock()
			notifier := &SlackNotifier{
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
