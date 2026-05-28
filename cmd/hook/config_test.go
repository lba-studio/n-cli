package hook

import (
	"testing"

	"github.com/lba-studio/n-cli/internal/config"
)

func stubHookConfig(t *testing.T, cfg config.Config) {
	t.Helper()
	original := loadHookConfig
	loadHookConfig = func() (config.Config, error) {
		return cfg, nil
	}
	t.Cleanup(func() {
		loadHookConfig = original
	})
}
