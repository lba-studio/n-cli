package notifier

import (
	"testing"

	"github.com/lba-studio/n-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCustomNotifierEntries(t *testing.T) {
	t.Run("legacy custom only", func(t *testing.T) {
		cfg := config.Config{
			Custom: &config.CustomConfig{
				Name:            "legacy",
				TargetUrl:       "https://api.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 1)
		assert.Equal(t, "legacy", entries[0].label)
		assert.Equal(t, cfg.Custom.TargetUrl, entries[0].cfg.TargetUrl)
	})

	t.Run("customs array only", func(t *testing.T) {
		cfg := config.Config{
			Customs: []config.CustomConfig{
				{
					Name:            "pagerduty",
					TargetUrl:       "https://api.example.com/webhook",
					PayloadTemplate: `{"message": "{{message}}"}`,
				},
				{
					TargetUrl:       "https://other.example.com/notify",
					PayloadTemplate: "Alert: {{message}}",
				},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 2)
		assert.Equal(t, "pagerduty", entries[0].label)
		assert.Equal(t, "custom[1]", entries[1].label)
	})

	t.Run("legacy custom prepended to customs", func(t *testing.T) {
		cfg := config.Config{
			Custom: &config.CustomConfig{
				TargetUrl:       "https://legacy.example.com/webhook",
				PayloadTemplate: `{"message": "{{message}}"}`,
			},
			Customs: []config.CustomConfig{
				{
					Name:            "monitoring",
					TargetUrl:       "https://monitoring.example.com/notify",
					PayloadTemplate: "Alert: {{message}}",
				},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 2)
		assert.Equal(t, "custom[0]", entries[0].label)
		assert.Equal(t, "monitoring", entries[1].label)
		assert.Equal(t, "https://legacy.example.com/webhook", entries[0].cfg.TargetUrl)
		assert.Equal(t, "https://monitoring.example.com/notify", entries[1].cfg.TargetUrl)
	})

	t.Run("empty when no custom config", func(t *testing.T) {
		entries := customNotifierEntries(config.Config{})
		assert.Empty(t, entries)
	})
}
