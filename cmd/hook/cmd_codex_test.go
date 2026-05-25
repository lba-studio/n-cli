package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatCodexMessage(t *testing.T) {
	tests := []struct {
		name    string
		payload codexHookPayload
		want    string
	}{
		{
			name: "permission request with tool",
			payload: codexHookPayload{
				HookEventName: "PermissionRequest",
				ToolName:      "shell",
			},
			want: "Codex needs approval for: shell",
		},
		{
			name: "permission request without tool",
			payload: codexHookPayload{
				HookEventName: "PermissionRequest",
			},
			want: "Codex needs approval for: unknown tool",
		},
		{
			name: "stop event",
			payload: codexHookPayload{
				HookEventName: "Stop",
			},
			want: "Codex agent finished",
		},
		{
			name: "unknown event",
			payload: codexHookPayload{
				HookEventName: "Custom",
			},
			want: "Codex hook: Custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatCodexMessage(tt.payload))
		})
	}
}

func TestHandleHookCodex(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})

	tests := []struct {
		name      string
		data      []byte
		wantMsg   string
		wantErr   bool
		wantNotif bool
	}{
		{
			name:      "valid permission request",
			data:      []byte(`{"hook_event_name":"PermissionRequest","tool_name":"shell"}`),
			wantMsg:   "Codex needs approval for: shell",
			wantNotif: true,
		},
		{
			name:      "valid stop event",
			data:      []byte(`{"hook_event_name":"Stop"}`),
			wantMsg:   "Codex agent finished",
			wantNotif: true,
		},
		{
			name:    "malformed JSON",
			data:    []byte(`not json`),
			wantErr: true,
		},
		{
			name:      "empty JSON",
			data:      []byte(`{}`),
			wantMsg:   "Codex hook: unknown",
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

			err := HandleHookCodex(tt.data)
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
