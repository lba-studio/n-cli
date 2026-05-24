package config

type DiscordConfig = WebhookConfig

type SlackConfig = WebhookConfig

type WebhookConfig struct {
	WebhookURL    string `mapstructure:"webhookUrl"`
	MessageFormat string `mapstructure:"messageFormat"`
}

type CustomConfig struct {
	PayloadTemplate string            `mapstructure:"payloadTemplate"`
	TargetUrl       string            `mapstructure:"targetUrl"`
	Method          string            `mapstructure:"method"`
	Headers         map[string]string `mapstructure:"headers"`
}

type SystemConfig struct {
	Disabled bool `mapstructure:"disabled"`
}

// Config struct to hold the configuration values
type Config struct {
	Discord *DiscordConfig `mapstructure:"discord"`
	Slack   *SlackConfig   `mapstructure:"slack"`
	Custom  *CustomConfig  `mapstructure:"custom"`
	System  *SystemConfig  `mapstructure:"system"`
}
