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
	assert.Equal(t, false, StringMapValue(readToggleConfig(t, path)["slack"])["disabled"])
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
	assert.Equal(t, false, StringMapValue(readToggleConfig(t, path)["system"])["disabled"])
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
	assert.Equal(t, false, StringMapValue(readToggleConfig(t, path)["custom"])["disabled"])
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

func toggle(t *testing.T, content, label string, disabled bool) string {
	t.Helper()
	path := writeToggleConfig(t, content)
	_, _, err := SetChannelDisabled(path, label, disabled)
	require.NoError(t, err)
	out, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(out)
}

func TestSetChannelDisabledEditsInPlace(t *testing.T) {
	const src = `# n-cli config

slack:   # team workspace
  webhookUrl: "https://slack.example.com"   # keep quoted

customs:
  # first entry
  - name: pagerduty
    targetUrl: https://pd.example.com

  - targetUrl: https://two.example.com
    disabled: false # trailing comment
discord:
    webhookUrl: https://discord.example.com
`

	assert.Equal(t, `# n-cli config

slack:   # team workspace
  disabled: true
  webhookUrl: "https://slack.example.com"   # keep quoted

customs:
  # first entry
  - name: pagerduty
    targetUrl: https://pd.example.com

  - targetUrl: https://two.example.com
    disabled: false # trailing comment
discord:
    webhookUrl: https://discord.example.com
`, toggle(t, src, "slack", true))

	assert.Equal(t, `# n-cli config

slack:   # team workspace
  webhookUrl: "https://slack.example.com"   # keep quoted

customs:
  # first entry
  - disabled: true
    name: pagerduty
    targetUrl: https://pd.example.com

  - targetUrl: https://two.example.com
    disabled: false # trailing comment
discord:
    webhookUrl: https://discord.example.com
`, toggle(t, src, "pagerduty", true))

	assert.Equal(t, `# n-cli config

slack:   # team workspace
  webhookUrl: "https://slack.example.com"   # keep quoted

customs:
  # first entry
  - name: pagerduty
    targetUrl: https://pd.example.com

  - targetUrl: https://two.example.com
    disabled: true # trailing comment
discord:
    webhookUrl: https://discord.example.com
`, toggle(t, src, "customs[1]", true))

	assert.Equal(t, `# n-cli config

slack:   # team workspace
  webhookUrl: "https://slack.example.com"   # keep quoted

customs:
  # first entry
  - name: pagerduty
    targetUrl: https://pd.example.com

  - targetUrl: https://two.example.com
    disabled: false # trailing comment
discord:
    disabled: true
    webhookUrl: https://discord.example.com
`, toggle(t, src, "discord", true), "keeps the block's own indentation")
}

func TestSetChannelDisabledValueForms(t *testing.T) {
	assert.Equal(t, "slack:\n  disabled: false\n  webhookUrl: x\n",
		toggle(t, "slack:\n  disabled: True\n  webhookUrl: x\n", "slack", false))
	assert.Equal(t, "slack: {disabled: true, webhookUrl: x}\n",
		toggle(t, "slack: {webhookUrl: x}\n", "slack", true))
	assert.Equal(t, "slack: {webhookUrl: x, disabled: false}\n",
		toggle(t, "slack: {webhookUrl: x, disabled: true}\n", "slack", false))
	assert.Equal(t, "slack:\r\n  disabled: true\r\n  webhookUrl: x\r\n",
		toggle(t, "slack:\r\n  webhookUrl: x\r\n", "slack", true))
	assert.Equal(t, "customs:\n  - {disabled: true, name: \"\u00e9t\u00e9\", targetUrl: x}\n",
		toggle(t, "customs:\n  - {name: \"\u00e9t\u00e9\", targetUrl: x}\n", "\u00e9t\u00e9", true))
}

func TestSetChannelDisabledSystemBlock(t *testing.T) {
	assert.Equal(t, "system:\n  disabled: true\n", toggle(t, "", "system", true))
	assert.Equal(t, "# comment only\nsystem:\n  disabled: true\n", toggle(t, "# comment only", "system", true))
	assert.Equal(t, "slack:\n  webhookUrl: x\nsystem:\n  disabled: true\n",
		toggle(t, "slack:\n  webhookUrl: x\n", "system", true))
	assert.Equal(t, "system: # desktop\n  disabled: true\nslack:\n  webhookUrl: x\n",
		toggle(t, "system: # desktop\nslack:\n  webhookUrl: x\n", "system", true))

	missing := filepath.Join(t.TempDir(), "nested", "config.yaml")
	_, changed, err := SetChannelDisabled(missing, "system", true)
	require.NoError(t, err)
	assert.True(t, changed)
	out, err := os.ReadFile(missing)
	require.NoError(t, err)
	assert.Equal(t, "system:\n  disabled: true\n", string(out))
}

func TestSetChannelDisabledMergeKey(t *testing.T) {
	const src = `base: &base
  disabled: true
customs:
  - <<: *base
    name: inherited
`
	assert.Equal(t, `base: &base
  disabled: true
customs:
  - disabled: false
    <<: *base
    name: inherited
`, toggle(t, src, "inherited", false))
}

func TestSetChannelDisabledRefusesUnsafeEdits(t *testing.T) {
	cases := map[string]struct {
		content, label, err string
	}{
		"alias":         {"base: &b\n  url: x\ncustoms:\n  - *b\n", "customs[0]", "YAML alias"},
		"mapping value": {"slack:\n  disabled:\n    x: 1\n", "slack", "must be true or false"},
		"second doc":    {"slack:\n  url: x\n---\nother: 1\n", "system", "refusing to write"},
		"empty block":   {"slack: {}\ndiscord:\n  url: x\n", "slack", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := writeToggleConfig(t, tc.content)
			_, _, err := SetChannelDisabled(path, tc.label, true)
			if tc.err == "" {
				require.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.err)
			out, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			assert.Equal(t, tc.content, string(out))
		})
	}
}
