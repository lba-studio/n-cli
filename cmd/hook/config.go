package hook

import (
	"github.com/lba-studio/n-cli/internal/config"
)

type hookAgent string

const (
	hookAgentCursor     hookAgent = config.HookAgentCursorKey
	hookAgentCodex      hookAgent = config.HookAgentCodexKey
	hookAgentClaudeCode hookAgent = config.HookAgentClaudeCodeKey
)

var loadHookConfig = config.GetConfig

func isHookEventIgnored(agent hookAgent, event string) (bool, error) {
	cfg, err := loadHookConfig()
	if err != nil {
		return false, err
	}
	if cfg.Hooks == nil {
		return false, nil
	}

	var agentCfg *config.HookAgentConfig
	switch agent {
	case hookAgentCursor:
		agentCfg = cfg.Hooks.Cursor
	case hookAgentCodex:
		agentCfg = cfg.Hooks.Codex
	case hookAgentClaudeCode:
		agentCfg = cfg.Hooks.ClaudeCode
	}
	if agentCfg == nil {
		return false, nil
	}

	for _, ignored := range agentCfg.IgnoredEvents {
		if event == ignored {
			return true, nil
		}
	}
	return false, nil
}
