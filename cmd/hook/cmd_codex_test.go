package hook

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lba-studio/n-cli/internal/config"
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
		{
			name: "stop with cwd shows project name",
			payload: codexHookPayload{
				HookEventName: "Stop",
				Cwd:           "/Users/user/n-cli",
			},
			want: "Codex [n-cli] agent finished",
		},
		{
			name: "permission request with cwd shows project name",
			payload: codexHookPayload{
				HookEventName: "PermissionRequest",
				ToolName:      "shell",
				Cwd:           "/Users/user/n-cli",
			},
			want: "Codex [n-cli] needs approval for: shell",
		},
		{
			name: "root cwd ignored",
			payload: codexHookPayload{
				HookEventName: "Stop",
				Cwd:           "/",
			},
			want: "Codex agent finished",
		},
		{
			name: "dot cwd ignored",
			payload: codexHookPayload{
				HookEventName: "Stop",
				Cwd:           ".",
			},
			want: "Codex agent finished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatCodexMessage(tt.payload))
		})
	}
}

func TestCodexAgentLabel(t *testing.T) {
	tests := []struct {
		name        string
		cwd         string
		want        string
		windowsOnly bool
	}{
		{name: "unix path", cwd: "/Users/user/my-project", want: "my-project"},
		{name: "windows path", cwd: `C:\Users\user\my-project`, want: "my-project", windowsOnly: true},
		{name: "empty", cwd: "", want: ""},
		{name: "dot", cwd: ".", want: ""},
		{name: "unix root", cwd: "/", want: ""},
		{name: "windows root", cwd: `\`, want: "", windowsOnly: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("Windows path semantics only apply on Windows")
			}
			assert.Equal(t, tt.want, codexAgentLabel(tt.cwd))
		})
	}
}

func TestCodexThreadName(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "session_index.jsonl")

	t.Run("returns empty when session_id is empty", func(t *testing.T) {
		assert.Equal(t, "", codexThreadNameFromPath("", indexPath))
	})

	t.Run("returns empty when file does not exist", func(t *testing.T) {
		assert.Equal(t, "", codexThreadNameFromPath("abc", filepath.Join(dir, "missing.jsonl")))
	})

	t.Run("returns thread name for matching session", func(t *testing.T) {
		require.NoError(t, os.WriteFile(indexPath, []byte(
			`{"id":"aaa","thread_name":"Fix auth bug","updated_at":"2026-01-01T00:00:00Z"}`+"\n"+
				`{"id":"bbb","thread_name":"Other session","updated_at":"2026-01-01T00:01:00Z"}`+"\n",
		), 0600))
		assert.Equal(t, "Fix auth bug", codexThreadNameFromPath("aaa", indexPath))
	})

	t.Run("last rename wins", func(t *testing.T) {
		require.NoError(t, os.WriteFile(indexPath, []byte(
			`{"id":"aaa","thread_name":"Old name","updated_at":"2026-01-01T00:00:00Z"}`+"\n"+
				`{"id":"aaa","thread_name":"New name","updated_at":"2026-01-01T00:01:00Z"}`+"\n",
		), 0600))
		assert.Equal(t, "New name", codexThreadNameFromPath("aaa", indexPath))
	})

	t.Run("returns empty for unmatched session", func(t *testing.T) {
		require.NoError(t, os.WriteFile(indexPath, []byte(
			`{"id":"aaa","thread_name":"Fix auth bug","updated_at":"2026-01-01T00:00:00Z"}`+"\n",
		), 0600))
		assert.Equal(t, "", codexThreadNameFromPath("zzz", indexPath))
	})
}

func TestHandleHookCodex(t *testing.T) {
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
			name:      "valid permission request",
			data:      []byte(`{"hook_event_name":"PermissionRequest","tool_name":"shell","cwd":"/Users/user/n-cli"}`),
			wantMsg:   "Codex [n-cli] needs approval for: shell",
			wantNotif: true,
		},
		{
			name:      "valid stop event",
			data:      []byte(`{"hook_event_name":"Stop","cwd":"/Users/user/n-cli"}`),
			wantMsg:   "Codex [n-cli] agent finished",
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

func TestHandleHookCodexIgnoredEvent(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{
		Hooks: &config.HooksConfig{
			Codex: &config.HookAgentConfig{
				IgnoredEvents: []string{"PermissionRequest", "Stop"},
			},
		},
	})

	var gotMsg string
	notify = func(msg string) error {
		gotMsg = msg
		return nil
	}

	output, err := HandleHookCodex([]byte(`{"hook_event_name":"Stop"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"continue":true}`, string(output))
	assert.Empty(t, gotMsg)
}

func TestHandleHookCodexNonIgnoredEvent(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{
		Hooks: &config.HooksConfig{
			Codex: &config.HookAgentConfig{
				IgnoredEvents: []string{"PermissionRequest"},
			},
		},
	})

	var gotMsg string
	notify = func(msg string) error {
		gotMsg = msg
		return nil
	}

	output, err := HandleHookCodex([]byte(`{"hook_event_name":"Stop"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"continue":true}`, string(output))
	assert.Equal(t, "Codex agent finished", gotMsg)
}

func TestHookCodexCommandOutput(t *testing.T) {
	originalNotify := notify
	t.Cleanup(func() {
		notify = originalNotify
	})
	stubHookConfig(t, config.Config{})

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
			input:      `{"hook_event_name":"Stop","cwd":"/Users/user/n-cli"}`,
			wantMsg:    "Codex [n-cli] agent finished",
			wantStdout: `{"continue":true}`,
		},
		{
			name:       "stop writes valid JSON when notification fails",
			input:      `{"hook_event_name":"Stop","cwd":"/Users/user/n-cli"}`,
			wantMsg:    "Codex [n-cli] agent finished",
			wantStdout: `{"continue":true}`,
			notifyErr:  errors.New("boom"),
			wantStderr: "n-cli hook error: boom\n",
		},
		{
			name:    "permission request writes no stdout",
			input:   `{"hook_event_name":"PermissionRequest","tool_name":"shell","cwd":"/Users/user/n-cli"}`,
			wantMsg: "Codex [n-cli] needs approval for: shell",
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
