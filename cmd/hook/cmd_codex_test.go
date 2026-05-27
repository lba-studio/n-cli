package hook

import (
	"bytes"
	"errors"
	"strings"
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

			_, err := HandleHookCodex(tt.data)
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

func TestHookCodexCommandOutput(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})

	tests := []struct {
		name       string
		input      string
		wantMsg    string
		wantStdout string
		notifyErr  error
		wantStderr string
	}{
		{
			name:       "stop writes valid JSON",
			input:      `{"hook_event_name":"Stop"}`,
			wantMsg:    "Codex agent finished",
			wantStdout: `{"continue":true}`,
		},
		{
			name:       "stop writes valid JSON when notification fails",
			input:      `{"hook_event_name":"Stop"}`,
			wantMsg:    "Codex agent finished",
			wantStdout: `{"continue":true}`,
			notifyErr:  errors.New("boom"),
			wantStderr: "n-cli hook error: boom\n",
		},
		{
			name:    "permission request writes no stdout",
			input:   `{"hook_event_name":"PermissionRequest","tool_name":"shell"}`,
			wantMsg: "Codex needs approval for: shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMsg string
			notify = func(msg string) error {
				gotMsg = msg
				return tt.notifyErr
			}

			cmd := NewHookCodexCmd()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.SetIn(strings.NewReader(tt.input))
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			require.NoError(t, cmd.Execute())
			assert.Equal(t, tt.wantStderr, stderr.String())
			assert.Equal(t, tt.wantMsg, gotMsg)
			if tt.wantStdout == "" {
				assert.Empty(t, stdout.String())
			} else {
				assert.JSONEq(t, tt.wantStdout, stdout.String())
			}
		})
	}
}
