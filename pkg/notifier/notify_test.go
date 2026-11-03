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

	t.Run("disabled entries skipped and labels stay stable", func(t *testing.T) {
		cfg := config.Config{
			Customs: []config.CustomConfig{
				{TargetUrl: "https://a.example.com", Disabled: true},
				{TargetUrl: "https://b.example.com"},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 1)
		assert.Equal(t, "custom[1]", entries[0].label)
		assert.Equal(t, "https://b.example.com", entries[0].cfg.TargetUrl)
	})

	t.Run("empty when no custom config", func(t *testing.T) {
		entries := customNotifierEntries(config.Config{})
		assert.Empty(t, entries)
	})

	t.Run("duplicate names get suffixed", func(t *testing.T) {
		cfg := config.Config{
			Custom: &config.CustomConfig{Name: "alerts", TargetUrl: "https://a.example.com"},
			Customs: []config.CustomConfig{
				{Name: "alerts", TargetUrl: "https://b.example.com"},
				{Name: "alerts", TargetUrl: "https://c.example.com"},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 3)
		assert.Equal(t, "alerts", entries[0].label)
		assert.Empty(t, entries[0].renamedFrom)
		assert.Equal(t, "alerts#2", entries[1].label)
		assert.Equal(t, "alerts", entries[1].renamedFrom)
		assert.Equal(t, "alerts#3", entries[2].label)
		assert.Equal(t, "alerts", entries[2].renamedFrom)
		assert.Equal(t, "https://c.example.com", entries[2].cfg.TargetUrl)
	})

	t.Run("names clashing with built-in channels get suffixed", func(t *testing.T) {
		cfg := config.Config{
			Customs: []config.CustomConfig{
				{Name: "slack", TargetUrl: "https://a.example.com"},
				{Name: "system", TargetUrl: "https://b.example.com"},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 2)
		assert.Equal(t, "slack#2", entries[0].label)
		assert.Equal(t, "slack", entries[0].renamedFrom)
		assert.Equal(t, "system#2", entries[1].label)
		assert.Equal(t, "system", entries[1].renamedFrom)
	})

	t.Run("name clashing with auto-generated label gets suffixed", func(t *testing.T) {
		cfg := config.Config{
			Customs: []config.CustomConfig{
				{TargetUrl: "https://a.example.com"},
				{Name: "custom[0]", TargetUrl: "https://b.example.com"},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 2)
		assert.Equal(t, "custom[0]", entries[0].label)
		assert.Equal(t, "custom[0]#2", entries[1].label)
		assert.Equal(t, "custom[0]", entries[1].renamedFrom)
	})

	t.Run("unique names are not renamed", func(t *testing.T) {
		cfg := config.Config{
			Customs: []config.CustomConfig{
				{Name: "a", TargetUrl: "https://a.example.com"},
				{Name: "b", TargetUrl: "https://b.example.com"},
			},
		}

		entries := customNotifierEntries(cfg)
		assert.Len(t, entries, 2)
		for _, e := range entries {
			assert.Empty(t, e.renamedFrom)
		}
	})
}
