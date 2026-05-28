package hook

import (
	"testing"

	"github.com/lba-studio/n-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatCursorMessage(t *testing.T) {
	tests := []struct {
		name    string
		payload cursorHookPayload
		want    string
	}{
		{
			name: "stop with status",
			payload: cursorHookPayload{
				HookEventName: "stop",
				Status:        "completed",
			},
			want: "Done: agent finished (status: completed)",
		},
		{
			name: "stop without status",
			payload: cursorHookPayload{
				HookEventName: "stop",
			},
			want: "Done: agent finished (status: unknown)",
		},
		{
			name: "sessionEnd is silent",
			payload: cursorHookPayload{
				HookEventName: "sessionEnd",
			},
			want: "",
		},
		{
			name: "unknown event",
			payload: cursorHookPayload{
				HookEventName: "customEvent",
			},
			want: "Cursor hook: customEvent",
		},
		{
			name:    "empty event",
			payload: cursorHookPayload{},
			want:    "Cursor hook: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatCursorMessage(tt.payload))
		})
	}
}

func TestHandleHookCursor(t *testing.T) {
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
			name:      "valid stop event",
			data:      []byte(`{"hook_event_name":"stop","status":"completed"}`),
			wantMsg:   "Done: agent finished (status: completed)",
			wantNotif: true,
		},
		{
			name:      "sessionEnd skips notification",
			data:      []byte(`{"hook_event_name":"sessionEnd"}`),
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
			wantMsg:   "Cursor hook: unknown",
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

			err := HandleHookCursor(tt.data)
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

func TestHandleHookCursorIgnoredEvent(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{
		Hooks: &config.HooksConfig{
			Cursor: &config.HookAgentConfig{
				IgnoredEvents: []string{"stop"},
			},
		},
	})

	var gotMsg string
	notify = func(msg string) error {
		gotMsg = msg
		return nil
	}

	err := HandleHookCursor([]byte(`{"hook_event_name":"stop","status":"completed"}`))
	require.NoError(t, err)
	assert.Empty(t, gotMsg)
}

func TestHandleHookCursorNonIgnoredEvent(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{
		Hooks: &config.HooksConfig{
			Cursor: &config.HookAgentConfig{
				IgnoredEvents: []string{"sessionEnd"},
			},
		},
	})

	var gotMsg string
	notify = func(msg string) error {
		gotMsg = msg
		return nil
	}

	err := HandleHookCursor([]byte(`{"hook_event_name":"stop","status":"completed"}`))
	require.NoError(t, err)
	assert.Equal(t, "Done: agent finished (status: completed)", gotMsg)
}
