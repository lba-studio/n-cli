package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHooksConfigUnmarshalSnakeCase(t *testing.T) {
	input := map[string]interface{}{
		"hooks": map[string]interface{}{
			"codex": map[string]interface{}{
				"setup":          true,
				"ignored_events": []interface{}{"PermissionRequest"},
			},
			"claude_code": map[string]interface{}{
				"setup":          true,
				"ignored_events": []interface{}{"Notification"},
			},
			"cursor": map[string]interface{}{
				"setup":          true,
				"ignored_events": []interface{}{"stop"},
			},
		},
	}

	var cfg Config
	require.NoError(t, mapstructure.Decode(input, &cfg))

	require.NotNil(t, cfg.Hooks)
	require.NotNil(t, cfg.Hooks.Codex)
	require.NotNil(t, cfg.Hooks.ClaudeCode)
	require.NotNil(t, cfg.Hooks.Cursor)
	assert.Equal(t, []string{"PermissionRequest"}, cfg.Hooks.Codex.IgnoredEvents)
	assert.Equal(t, []string{"Notification"}, cfg.Hooks.ClaudeCode.IgnoredEvents)
	assert.Equal(t, []string{"stop"}, cfg.Hooks.Cursor.IgnoredEvents)
}

func TestHooksConfigMarshalSnakeCase(t *testing.T) {
	cfg := Config{
		Hooks: &HooksConfig{
			Codex: &HookAgentConfig{
				Setup:         true,
				IgnoredEvents: []string{"PermissionRequest"},
			},
			ClaudeCode: &HookAgentConfig{
				Setup:         true,
				IgnoredEvents: []string{"Notification"},
			},
			Cursor: &HookAgentConfig{
				Setup:         true,
				IgnoredEvents: []string{"stop"},
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(data), "claude_code:")
	assert.Contains(t, string(data), "ignored_events:")
	assert.NotContains(t, string(data), "ClaudeCode")
	assert.NotContains(t, string(data), "IgnoredEvents")
}
