package hook

import (
	"testing"

	"github.com/lba-studio/n-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatClaudeCodeMessage(t *testing.T) {
	tests := []struct {
		name    string
		payload claudeCodeHookPayload
		want    string
	}{
		{
			name: "notification permission prompt",
			payload: claudeCodeHookPayload{
				HookEventName:    "Notification",
				NotificationType: "permission_prompt",
				Message:          "approve tool use?",
			},
			want: "Claude Code needs approval: approve tool use?",
		},
		{
			name: "notification permission prompt without message",
			payload: claudeCodeHookPayload{
				HookEventName:    "Notification",
				NotificationType: "permission_prompt",
			},
			want: "Claude Code needs approval: permission needed",
		},
		{
			name: "notification other type",
			payload: claudeCodeHookPayload{
				HookEventName:    "Notification",
				NotificationType: "idle_prompt",
				Message:          "still there?",
			},
			want: "Claude Code: still there?",
		},
		{
			name: "notification other type without message",
			payload: claudeCodeHookPayload{
				HookEventName:    "Notification",
				NotificationType: "idle_prompt",
			},
			want: "Claude Code: notification",
		},
		{
			name: "stop event",
			payload: claudeCodeHookPayload{
				HookEventName: "Stop",
			},
			want: "Claude Code agent finished",
		},
		{
			name: "unknown event",
			payload: claudeCodeHookPayload{
				HookEventName: "Custom",
			},
			want: "Claude Code hook: Custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatClaudeCodeMessage(tt.payload))
		})
	}
}

func TestHandleHookClaudeCode(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{})

	tests := []struct {
		name      string
		data      []byte
		wantMsg   string
		wantErr   bool
		wantNotif bool
	}{
		{
			name:      "valid permission prompt",
			data:      []byte(`{"hook_event_name":"Notification","notification_type":"permission_prompt","message":"approve?"}`),
			wantMsg:   "Claude Code needs approval: approve?",
			wantNotif: true,
		},
		{
			name:      "valid stop event",
			data:      []byte(`{"hook_event_name":"Stop"}`),
			wantMsg:   "Claude Code agent finished",
			wantNotif: true,
		},
		{
			name:      "cursor stop event is ignored",
			data:      []byte(`{"hook_event_name":"stop","status":"completed","cursor_version":"3.5.33"}`),
			wantNotif: false,
		},
		{
			name:      "cursor payload with cursor_version is ignored",
			data:      []byte(`{"hook_event_name":"Notification","notification_type":"permission_prompt","message":"approve?","cursor_version":"3.5.33"}`),
			wantNotif: false,
		},
		{
			name:    "malformed JSON",
			data:    []byte(`not json`),
			wantErr: true,
		},
		{
			name:      "empty JSON",
			data:      []byte(`{}`),
			wantMsg:   "Claude Code hook: unknown",
			wantNotif: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMsg string
			notify = func(msg string) error {
				gotMsg = msg
				return nil
			}

			err := HandleHookClaudeCode(tt.data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.wantNotif {
				assert.Equal(t, tt.wantMsg, gotMsg)
			} else {
				assert.Empty(t, gotMsg)
			}
		})
	}
}

func TestHandleHookClaudeCodeIgnoredEvent(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{
		Hooks: &config.HooksConfig{
			ClaudeCode: &config.HookAgentConfig{
				IgnoredEvents: []string{"Notification"},
			},
		},
	})

	var gotMsg string
	notify = func(msg string) error {
		gotMsg = msg
		return nil
	}

	err := HandleHookClaudeCode([]byte(`{"hook_event_name":"Notification","notification_type":"permission_prompt","message":"approve?"}`))
	require.NoError(t, err)
	assert.Empty(t, gotMsg)
}

func TestHandleHookClaudeCodeNonIgnoredEvent(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{
		Hooks: &config.HooksConfig{
			ClaudeCode: &config.HookAgentConfig{
				IgnoredEvents: []string{"Notification"},
			},
		},
	})

	var gotMsg string
	notify = func(msg string) error {
		gotMsg = msg
		return nil
	}

	err := HandleHookClaudeCode([]byte(`{"hook_event_name":"Stop"}`))
	require.NoError(t, err)
	assert.Equal(t, "Claude Code agent finished", gotMsg)
}
