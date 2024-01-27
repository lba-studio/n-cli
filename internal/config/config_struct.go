package config

type DiscordConfig = WebhookConfig

type SlackConfig = WebhookConfig

type WebhookConfig struct {
	WebhookURL    string `mapstructure:"webhookUrl"`
	MessageFormat string `mapstructure:"messageFormat"`
}

// Config struct to hold the configuration values
type Config struct {
	Discord *DiscordConfig `mapstructure:"discord"`
	Slack   *SlackConfig   `mapstructure:"slack"`
}
