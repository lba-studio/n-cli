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

func TestDiscordNotifier(t *testing.T) {
	testRestyClient := resty.New()
	httpmock.ActivateNonDefault(testRestyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	defaultConfig := config.Config{
		Discord: &config.DiscordConfig{
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
				responder := httpmock.NewStringResponder(200, "")
				httpmock.RegisterResponder("POST", defaultConfig.Discord.WebhookURL, responder)
			},
		},
		{
			name: "do not run when discord has no config",
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(config.Config{}, nil)
			},
			wantErr:              ErrMissingDiscordConfig,
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
			name:    "sad path - fail to call discord",
			wantErr: errors.New("failed to call Discord: {\"message\": \"Yo this is broken\"}"),
			doMock: func() {
				mockConfigurer.On("GetConfig").Return(defaultConfig, nil)
				responder := httpmock.NewStringResponder(400, `{"message": "Yo this is broken"}`)
				httpmock.RegisterResponder("POST", defaultConfig.Discord.WebhookURL, responder)
			},
		},
	}

	for _, tc := range testCases {
		httpmock.Reset()
		mockConfigurer = config.NewMockConfigurer(t)
		t.Run(tc.name, func(t *testing.T) {
			tc.doMock()
			notifier := &DiscordNotifier{
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
