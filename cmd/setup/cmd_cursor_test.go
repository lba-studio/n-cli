package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSetupCursorCmd(t *testing.T) {
	c := NewSetupCursorCmd()
	require.NotNil(t, c)
	assert.Equal(t, "cursor", c.Name())
	assert.NotEmpty(t, c.Long)
	f := c.Flags().Lookup("force")
	require.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

func TestMergeHookEntry(t *testing.T) {
	cmd := hookCommand
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
				map[string]interface{}{"command": "other-script.sh"},
			},
			command:  cmd,
			wantLen:  2,
			wantSame: false,
		},
		{
			name: "slice with our command already",
			existing: []interface{}{
				map[string]interface{}{"command": cmd},
			},
			command:  cmd,
			wantLen:  1,
			wantSame: true,
		},
		{
			name: "slice with our command and others",
			existing: []interface{}{
				map[string]interface{}{"command": "other.sh"},
				map[string]interface{}{"command": cmd},
			},
			command:  cmd,
			wantLen:  2,
			wantSame: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeHookEntry(tt.existing, tt.command)
			assert.Len(t, got, tt.wantLen)
			var hasCommand bool
			for _, e := range got {
				if m, ok := e.(map[string]interface{}); ok {
					if c, _ := m["command"].(string); c == tt.command {
						hasCommand = true
						break
					}
				}
				if m, ok := e.(map[string]string); ok {
					if m["command"] == tt.command {
						hasCommand = true
						break
					}
				}
			}
			if tt.wantLen >= 1 {
				assert.True(t, hasCommand, "result should contain command %q", tt.command)
			}
			if tt.wantSame && tt.existing != nil {
				if existingSlice, ok := tt.existing.([]interface{}); ok {
					assert.Same(t, &existingSlice[0], &got[0])
				}
			}
		})
	}
}

func TestMergeHooksJSON_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")

	err := mergeHooksJSON(path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	assert.Equal(t, float64(1), root["version"])
	hooks, ok := root["hooks"].(map[string]interface{})
	require.True(t, ok)
	for _, event := range []string{"stop", "sessionEnd"} {
		entries, ok := hooks[event].([]interface{})
		require.True(t, ok, "hooks[%s] should be slice", event)
		require.Len(t, entries, 1)
		entry, ok := entries[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, hookCommand, entry["command"])
	}
}

func TestMergeHooksJSON_ExistingEmptyObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0644))

	err := mergeHooksJSON(path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	assert.Equal(t, float64(1), root["version"])
	hooks, ok := root["hooks"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, hookCommand, hooks["stop"].([]interface{})[0].(map[string]interface{})["command"])
	assert.Equal(t, hookCommand, hooks["sessionEnd"].([]interface{})[0].(map[string]interface{})["command"])
}

func TestMergeHooksJSON_ExistingWithHooksNoVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"stop": []interface{}{map[string]interface{}{"command": "existing.sh"}},
		},
	}
	raw, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(path, raw, 0644))

	err := mergeHooksJSON(path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	assert.Equal(t, float64(1), root["version"])
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks["stop"].([]interface{})
	require.Len(t, stopEntries, 2)
	assert.Equal(t, "existing.sh", stopEntries[0].(map[string]interface{})["command"])
	assert.Equal(t, hookCommand, stopEntries[1].(map[string]interface{})["command"])
	sessionEndEntries := hooks["sessionEnd"].([]interface{})
	require.Len(t, sessionEndEntries, 1)
	assert.Equal(t, hookCommand, sessionEndEntries[0].(map[string]interface{})["command"])
}

func TestMergeHooksJSON_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	err := mergeHooksJSON(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse existing")
}

func TestMergeHooksJSON_SecondCallPreservesHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hooks.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	existing := map[string]interface{}{
		"hooks": map[string]interface{}{},
	}
	raw, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(path, raw, 0644))

	require.NoError(t, mergeHooksJSON(path))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var root map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &root))
	hooks := root["hooks"].(map[string]interface{})
	stopEntries := hooks["stop"].([]interface{})
	require.Len(t, stopEntries, 1)
	assert.Equal(t, hookCommand, stopEntries[0].(map[string]interface{})["command"])
}
