package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const toggleTestConfig = `discord:
  webhookUrl: https://discord.example.com
slack:
  webhookUrl: https://slack.example.com
  disabled: true
custom:
  name: legacy
  targetUrl: https://legacy.example.com
customs:
  - name: pagerduty
    targetUrl: https://pd.example.com
  - targetUrl: https://unnamed.example.com
  - name: ops
    targetUrl: https://ops1.example.com
  - name: ops
    targetUrl: https://ops2.example.com
  - name: discord
    targetUrl: https://fake-discord.example.com
hooks:
  codex:
    setup: true
`

func writeToggleConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

func readToggleConfig(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	root, err := ReadConfigMap(path)
	require.NoError(t, err)
	return root
}

func customsAt(t *testing.T, root map[string]interface{}, i int) map[string]interface{} {
	t.Helper()
	return StringMapValue(root["customs"].([]interface{})[i])
}

func TestSetChannelDisabledBuiltins(t *testing.T) {
	path := writeToggleConfig(t, toggleTestConfig)

	resolved, changed, err := SetChannelDisabled(path, "discord", true)
	require.NoError(t, err)
	assert.Equal(t, "discord", resolved)
	assert.True(t, changed)
	root := readToggleConfig(t, path)
	assert.Equal(t, true, StringMapValue(root["discord"])["disabled"])
	assert.Equal(t, "https://discord.example.com", StringMapValue(root["discord"])["webhookUrl"])
	assert.NotContains(t, customsAt(t, root, 4), "disabled", "built-in name must win over a custom entry")

	_, changed, err = SetChannelDisabled(path, "discord", true)
	require.NoError(t, err)
	assert.False(t, changed)

	_, changed, err = SetChannelDisabled(path, "slack", false)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.NotContains(t, StringMapValue(readToggleConfig(t, path)["slack"]), "disabled")
}

func TestSetChannelDisabledSystem(t *testing.T) {
	path := writeToggleConfig(t, toggleTestConfig)

	_, changed, err := SetChannelDisabled(path, "system", false)
	require.NoError(t, err)
	assert.False(t, changed, "system is enabled by default")
	assert.NotContains(t, readToggleConfig(t, path), "system")

	_, changed, err = SetChannelDisabled(path, "system", true)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, true, StringMapValue(readToggleConfig(t, path)["system"])["disabled"])

	_, changed, err = SetChannelDisabled(path, "system", false)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.NotContains(t, StringMapValue(readToggleConfig(t, path)["system"]), "disabled")
}

func TestSetChannelDisabledCustoms(t *testing.T) {
	path := writeToggleConfig(t, toggleTestConfig)

	resolved, _, err := SetChannelDisabled(path, "customs[1]", true)
	require.NoError(t, err)
	assert.Equal(t, "customs[1]", resolved)

	resolved, _, err = SetChannelDisabled(path, "pagerduty", true)
	require.NoError(t, err)
	assert.Equal(t, "customs[0] (pagerduty)", resolved)

	resolved, _, err = SetChannelDisabled(path, "custom", true)
	require.NoError(t, err)
	assert.Equal(t, "custom (legacy)", resolved)

	root := readToggleConfig(t, path)
	assert.Equal(t, true, customsAt(t, root, 0)["disabled"])
	assert.Equal(t, true, customsAt(t, root, 1)["disabled"])
	assert.NotContains(t, customsAt(t, root, 2), "disabled")
	assert.Equal(t, true, StringMapValue(root["custom"])["disabled"])
	assert.Equal(t, true, StringMapValue(StringMapValue(root["hooks"])["codex"])["setup"], "unrelated keys survive")

	_, _, err = SetChannelDisabled(path, "legacy", false)
	require.NoError(t, err)
	assert.NotContains(t, StringMapValue(readToggleConfig(t, path)["custom"]), "disabled")
}

func TestSetChannelDisabledErrors(t *testing.T) {
	path := writeToggleConfig(t, toggleTestConfig)
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	_, _, err = SetChannelDisabled(path, "ops", true)
	assert.ErrorContains(t, err, `"ops" is ambiguous; use one of: customs[2], customs[3]`)

	_, _, err = SetChannelDisabled(path, "customs[5]", true)
	assert.ErrorContains(t, err, "customs[5] does not exist")

	_, _, err = SetChannelDisabled(path, "nope", true)
	assert.ErrorContains(t, err, `no channel matches "nope"`)

	_, _, err = SetChannelDisabled(path, "", true)
	assert.Error(t, err)

	empty := writeToggleConfig(t, "")
	_, _, err = SetChannelDisabled(empty, "discord", true)
	assert.ErrorContains(t, err, `"discord" is not configured`)
	_, _, err = SetChannelDisabled(empty, "custom", true)
	assert.ErrorContains(t, err, `"custom" is not configured`)

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after), "failed toggles must not rewrite the file")
}
