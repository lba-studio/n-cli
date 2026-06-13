package config

type DiscordConfig = WebhookConfig

type SlackConfig = WebhookConfig

type WebhookConfig struct {
	WebhookURL    string `mapstructure:"webhookUrl"`
	MessageFormat string `mapstructure:"messageFormat"`
}

type CustomConfig struct {
	Name            string            `mapstructure:"name" yaml:"name,omitempty"`
	PayloadTemplate string            `mapstructure:"payloadTemplate"`
	TargetUrl       string            `mapstructure:"targetUrl"`
	Method          string            `mapstructure:"method"`
	Headers         map[string]string `mapstructure:"headers"`
}

type SystemConfig struct {
	Disabled bool `mapstructure:"disabled"`
}

const (
	HooksKey               = "hooks"
	HookAgentCodexKey      = "codex"
	HookAgentClaudeCodeKey = "claude_code"
	HookAgentCursorKey     = "cursor"
	HookSetupKey           = "setup"
	HookIgnoredEventsKey   = "ignored_events"
)

type HookAgentConfig struct {
	Setup         bool     `mapstructure:"setup" yaml:"setup"`
	IgnoredEvents []string `mapstructure:"ignored_events" yaml:"ignored_events,omitempty"`
}

type HooksConfig struct {
	Codex      *HookAgentConfig `mapstructure:"codex" yaml:"codex,omitempty"`
	ClaudeCode *HookAgentConfig `mapstructure:"claude_code" yaml:"claude_code,omitempty"`
	Cursor     *HookAgentConfig `mapstructure:"cursor" yaml:"cursor,omitempty"`
}

// Config struct to hold the configuration values
type Config struct {
	Discord *DiscordConfig `mapstructure:"discord" yaml:"discord,omitempty"`
	Slack   *SlackConfig   `mapstructure:"slack" yaml:"slack,omitempty"`
	Custom  *CustomConfig   `mapstructure:"custom" yaml:"custom,omitempty"`
	Customs []CustomConfig  `mapstructure:"customs" yaml:"customs,omitempty"`
	System  *SystemConfig   `mapstructure:"system" yaml:"system,omitempty"`
	Hooks   *HooksConfig   `mapstructure:"hooks" yaml:"hooks,omitempty"`
}
