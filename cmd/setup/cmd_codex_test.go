package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSetupCodexCmd(t *testing.T) {
	c := NewSetupCodexCmd()
	require.NotNil(t, c)
	assert.Equal(t, "codex", c.Name())
	assert.NotEmpty(t, c.Long)
}

func TestMergeCodexHookEvent(t *testing.T) {
	cmd := codexHookCommand
	tests := []struct {
		name     string
		existing interface{}
		command  string
		wantLen  int
		wantSame bool
	}{
		{
			name:     "nil existing",
			existing: nil,
			command:  cmd,
			wantLen:  1,
			wantSame: false,
		},
		{
			name:     "non-slice existing",
			existing: "not a slice",
			command:  cmd,
			wantLen:  1,
			wantSame: false,
		},
		{
			name:     "empty slice",
			existing: []interface{}{},
			command:  cmd,
			wantLen:  1,
			wantSame: false,
		},
		{
			name: "slice without our command",
			existing: []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "other-hook",
						},
					},
				},
			},
			command:  cmd,
			wantLen:  2,
			wantSame: false,
		},
		{
			name: "slice with our command already",
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
			wantLen:  1,
			wantSame: true,
		},
		{
			name: "slice with our command and others",
			existing: []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "other-hook",
						},
					},
				},
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
			wantLen:  2,
			wantSame: true,
		},
		{
			name: "replaces old shell hook",
			existing: []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "/home/user/.codex/hooks/n-cli-notify.sh",
						},
					},
				},
			},
			command:  cmd,
			wantLen:  1,
			wantSame: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCodexHookEvent(tt.existing, tt.command)
			assert.Len(t, got, tt.wantLen)
			var hasCommand bool
			for _, g := range got {
				group, ok := g.(map[string]interface{})
				require.True(t, ok)
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

func assertCodexHookEvent(t *testing.T, hooks map[string]interface{}, event, command string) {
	t.Helper()
	entries, ok := hooks[event].([]interface{})
	require.True(t, ok, "hooks[%s] should be slice", event)
	require.NotEmpty(t, entries)

	var hasCommand bool
	for _, g := range entries {
		group, ok := g.(map[string]interface{})
		require.True(t, ok)
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

func TestMergeCodexHooksJSON_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")

	err := mergeCodexHooksJSON(path, codexHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks, ok := root["hooks"].(map[string]interface{})
	require.True(t, ok)
	for _, event := range codexHookEvents {
		assertCodexHookEvent(t, hooks, event, codexHookCommand)
	}
}

func TestMergeCodexHooksJSON_ExistingEmptyObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0644))

	err := mergeCodexHooksJSON(path, codexHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks, ok := root["hooks"].(map[string]interface{})
	require.True(t, ok)
	for _, event := range codexHookEvents {
		assertCodexHookEvent(t, hooks, event, codexHookCommand)
	}
}

func TestMergeCodexHooksJSON_ExistingWithHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"Stop": []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "existing-hook",
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(path, raw, 0644))

	err := mergeCodexHooksJSON(path, codexHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks["Stop"].([]interface{})
	require.Len(t, stopEntries, 2)
	assert.Equal(t, "existing-hook", stopEntries[0].(map[string]interface{})["hooks"].([]interface{})[0].(map[string]interface{})["command"])
	assertCodexHookEvent(t, hooks, "Stop", codexHookCommand)
	assertCodexHookEvent(t, hooks, "PermissionRequest", codexHookCommand)
}

func TestMergeCodexHooksJSON_ReplacesOldShellHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"Stop": []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "/home/user/.codex/hooks/n-cli-notify.sh",
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(path, raw, 0644))

	err := mergeCodexHooksJSON(path, codexHookCommand)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks["Stop"].([]interface{})
	require.Len(t, stopEntries, 1)
	assertCodexHookEvent(t, hooks, "Stop", codexHookCommand)
}

func TestMergeCodexHooksJSON_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	err := mergeCodexHooksJSON(path, codexHookCommand)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse existing")
}

func TestMergeCodexHooksJSON_SecondCallPreservesHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"hooks":{}}`), 0644))

	require.NoError(t, mergeCodexHooksJSON(path, codexHookCommand))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks["Stop"].([]interface{})
	require.Len(t, stopEntries, 1)

	require.NoError(t, mergeCodexHooksJSON(path, codexHookCommand))
	data, err = os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &root))
	hooks = root["hooks"].(map[string]interface{})
	stopEntries = hooks["Stop"].([]interface{})
	require.Len(t, stopEntries, 1)
	assertCodexHookEvent(t, hooks, "Stop", codexHookCommand)
	assertCodexHookEvent(t, hooks, "PermissionRequest", codexHookCommand)
}
