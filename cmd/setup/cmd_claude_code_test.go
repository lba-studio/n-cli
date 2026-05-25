package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSetupClaudeCodeCmd(t *testing.T) {
	c := NewSetupClaudeCodeCmd()
	require.NotNil(t, c)
	assert.Equal(t, "claude-code", c.Name())
	assert.NotEmpty(t, c.Long)
}

func TestMergeClaudeCodeHookEvent(t *testing.T) {
	cmd := claudeCodeHookCommand
	tests := []struct {
		name     string
		existing interface{}
		command  string
		matcher  string
		wantLen  int
		wantSame bool
	}{
		{
			name:     "nil existing",
			existing: nil,
			command:  cmd,
			matcher:  "permission_prompt",
			wantLen:  1,
			wantSame: false,
		},
		{
			name:     "non-slice existing",
			existing: "not a slice",
			command:  cmd,
			matcher:  "",
			wantLen:  1,
			wantSame: false,
		},
		{
			name: "slice without our command",
			existing: []interface{}{
				map[string]interface{}{
					"matcher": "permission_prompt",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "other-hook",
						},
					},
				},
			},
			command:  cmd,
			matcher:  "permission_prompt",
			wantLen:  2,
			wantSame: false,
		},
		{
			name: "slice with our command and matching matcher",
			existing: []interface{}{
				map[string]interface{}{
					"matcher": "permission_prompt",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": cmd,
						},
					},
				},
			},
			command:  cmd,
			matcher:  "permission_prompt",
			wantLen:  1,
			wantSame: true,
		},
		{
			name: "slice with our command but different matcher",
			existing: []interface{}{
				map[string]interface{}{
					"matcher": "idle_prompt",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": cmd,
						},
					},
				},
			},
			command:  cmd,
			matcher:  "permission_prompt",
			wantLen:  2,
			wantSame: false,
		},
		{
			name: "stop event with our command already",
			existing: []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": cmd,
						},
					},
				},
			},
			command:  cmd,
			matcher:  "",
			wantLen:  1,
			wantSame: true,
		},
		{
			name: "replaces old shell hook",
			existing: []interface{}{
				map[string]interface{}{
					"matcher": "permission_prompt",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "/home/user/.claude/hooks/n-cli-notify.sh",
						},
					},
				},
			},
			command:  cmd,
			matcher:  "permission_prompt",
			wantLen:  1,
			wantSame: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeClaudeCodeHookEvent(tt.existing, tt.command, tt.matcher)
			assert.Len(t, got, tt.wantLen)
			var hasCommand bool
			for _, g := range got {
				group, ok := g.(map[string]interface{})
				require.True(t, ok)
				if !claudeCodeMatcherMatches(group["matcher"], tt.matcher) {
					continue
				}
				hooksList, ok := group["hooks"].([]interface{})
				require.True(t, ok)
				for _, h := range hooksList {
					hook, ok := h.(map[string]interface{})
					require.True(t, ok)
					if c, _ := hook["command"].(string); c == tt.command {
						hasCommand = true
						assert.Equal(t, "command", hook["type"])
						break
					}
				}
			}
			if tt.wantLen >= 1 {
				assert.True(t, hasCommand, "result should contain command %q", tt.command)
			}
		})
	}
}

func assertClaudeCodeHookEvent(t *testing.T, hooks map[string]interface{}, event, command, matcher string) {
	t.Helper()
	entries, ok := hooks[event].([]interface{})
	require.True(t, ok, "hooks[%s] should be slice", event)
	require.NotEmpty(t, entries)

	var hasCommand bool
	for _, g := range entries {
		group, ok := g.(map[string]interface{})
		require.True(t, ok)
		if matcher != "" {
			assert.Equal(t, matcher, group["matcher"])
		}
		hooksList, ok := group["hooks"].([]interface{})
		require.True(t, ok)
		for _, h := range hooksList {
			hook, ok := h.(map[string]interface{})
			require.True(t, ok)
			if c, _ := hook["command"].(string); c == command {
				hasCommand = true
				assert.Equal(t, "command", hook["type"])
				break
			}
		}
	}
	assert.True(t, hasCommand, "hooks[%s] should contain command %q", event, command)
}

func TestMergeClaudeCodeSettings_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, claudeCodeSettings)

	err := mergeClaudeCodeSettings(path, claudeCodeHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks, ok := root["hooks"].(map[string]interface{})
	require.True(t, ok)
	assertClaudeCodeHookEvent(t, hooks, claudeCodeStopEvent, claudeCodeHookCommand, "")
	assertClaudeCodeHookEvent(t, hooks, claudeCodeNotifyEvent, claudeCodeHookCommand, "permission_prompt")
}

func TestMergeClaudeCodeSettings_PreservesOtherSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, claudeCodeSettings)
	existing := map[string]interface{}{
		"permissions": map[string]interface{}{
			"defaultMode": "default",
		},
	}
	raw, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(path, raw, 0644))

	err := mergeClaudeCodeSettings(path, claudeCodeHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	permissions, ok := root["permissions"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "default", permissions["defaultMode"])
	hooks, ok := root["hooks"].(map[string]interface{})
	require.True(t, ok)
	assertClaudeCodeHookEvent(t, hooks, claudeCodeStopEvent, claudeCodeHookCommand, "")
	assertClaudeCodeHookEvent(t, hooks, claudeCodeNotifyEvent, claudeCodeHookCommand, "permission_prompt")
}

func TestMergeClaudeCodeSettings_ReplacesOldShellHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, claudeCodeSettings)
	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			claudeCodeStopEvent: []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "/home/user/.claude/hooks/n-cli-notify.sh",
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(path, raw, 0644))

	err := mergeClaudeCodeSettings(path, claudeCodeHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks[claudeCodeStopEvent].([]interface{})
	require.Len(t, stopEntries, 1)
	assertClaudeCodeHookEvent(t, hooks, claudeCodeStopEvent, claudeCodeHookCommand, "")
	assertClaudeCodeHookEvent(t, hooks, claudeCodeNotifyEvent, claudeCodeHookCommand, "permission_prompt")
}

func TestMergeClaudeCodeSettings_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, claudeCodeSettings)
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	err := mergeClaudeCodeSettings(path, claudeCodeHookCommand)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse existing")
}

func TestMergeClaudeCodeSettings_SecondCallPreservesHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, claudeCodeSettings)
	require.NoError(t, os.WriteFile(path, []byte(`{"hooks":{}}`), 0644))

	require.NoError(t, mergeClaudeCodeSettings(path, claudeCodeHookCommand))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks[claudeCodeStopEvent].([]interface{})
	require.Len(t, stopEntries, 1)

	require.NoError(t, mergeClaudeCodeSettings(path, claudeCodeHookCommand))
	data, err = os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &root))
	hooks = root["hooks"].(map[string]interface{})
	stopEntries = hooks[claudeCodeStopEvent].([]interface{})
	require.Len(t, stopEntries, 1)
	assertClaudeCodeHookEvent(t, hooks, claudeCodeStopEvent, claudeCodeHookCommand, "")
	assertClaudeCodeHookEvent(t, hooks, claudeCodeNotifyEvent, claudeCodeHookCommand, "permission_prompt")
}
