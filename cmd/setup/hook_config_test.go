package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIgnoredEvents(t *testing.T) {
	assert.Equal(t, []string{"PermissionRequest", "Future Event", "Custom"}, parseIgnoredEvents("PermissionRequest, Future Event ,Custom"))
	assert.Nil(t, parseIgnoredEvents(""))
	assert.Nil(t, parseIgnoredEvents(" , "))
}

func TestUpdateHookSetupConfigReplaceAndClearIgnoredEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`system:
  disabled: true
hooks:
  codex:
    setup: true
    ignored_events:
      - Stop
`), 0644))

	require.NoError(t, updateHookSetupConfig(path, codexHookSetupAgent, hookSetupOptions{
		ignoredEvents:        "PermissionRequest, Future Event",
		ignoredEventsChanged: true,
	}))

	root, err := readConfigMap(path)
	require.NoError(t, err)
	system := root["system"].(map[string]interface{})
	assert.Equal(t, true, system["disabled"])
	codex := root["hooks"].(map[string]interface{})["codex"].(map[string]interface{})
	assert.Equal(t, true, codex["setup"])
	assert.Equal(t, []interface{}{"PermissionRequest", "Future Event"}, codex["ignored_events"])

	require.NoError(t, updateHookSetupConfig(path, codexHookSetupAgent, hookSetupOptions{
		ignoredEvents:        "",
		ignoredEventsChanged: true,
	}))

	root, err = readConfigMap(path)
	require.NoError(t, err)
	codex = root["hooks"].(map[string]interface{})["codex"].(map[string]interface{})
	assert.Equal(t, true, codex["setup"])
	assert.NotContains(t, codex, "ignored_events")
}

func TestUpdateHookSetupConfigStoresArbitraryEventNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	require.NoError(t, updateHookSetupConfig(path, cursorHookSetupAgent, hookSetupOptions{
		ignoredEvents:        "not-a-real-event",
		ignoredEventsChanged: true,
	}))

	root, err := readConfigMap(path)
	require.NoError(t, err)
	cursor := root["hooks"].(map[string]interface{})["cursor"].(map[string]interface{})
	assert.Equal(t, []interface{}{"not-a-real-event"}, cursor["ignored_events"])
}

func TestUpdateHookSetupConfigFirstSetup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	require.NoError(t, updateHookSetupConfig(path, claudeCodeHookSetupAgent, hookSetupOptions{}))

	root, err := readConfigMap(path)
	require.NoError(t, err)
	claudeCode := root["hooks"].(map[string]interface{})["claude_code"].(map[string]interface{})
	assert.Equal(t, true, claudeCode["setup"])
	assert.NotContains(t, claudeCode, "ignored_events")
}

func TestUpdateHookSetupConfigExistingAgentKeepsIgnoredEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`hooks:
  cursor:
    setup: false
    ignored_events:
      - stop
`), 0644))

	require.NoError(t, updateHookSetupConfig(path, cursorHookSetupAgent, hookSetupOptions{}))

	root, err := readConfigMap(path)
	require.NoError(t, err)
	cursor := root["hooks"].(map[string]interface{})["cursor"].(map[string]interface{})
	assert.Equal(t, true, cursor["setup"])
	assert.Equal(t, []interface{}{"stop"}, cursor["ignored_events"])
}

func TestSetupCommandsExposeIgnoredEventsFlag(t *testing.T) {
	assert.NotNil(t, NewSetupCursorCmd().Flags().Lookup("ignored-events"))
	assert.NotNil(t, NewSetupCodexCmd().Flags().Lookup("ignored-events"))
	assert.NotNil(t, NewSetupClaudeCodeCmd().Flags().Lookup("ignored-events"))
}
