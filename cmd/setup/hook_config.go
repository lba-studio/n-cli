package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lba-studio/n-cli/internal/config"
)

type hookSetupAgent struct {
	configKey string
}

type hookSetupOptions struct {
	ignoredEvents        string
	ignoredEventsChanged bool
}

var (
	cursorHookSetupAgent = hookSetupAgent{
		configKey: config.HookAgentCursorKey,
	}
	codexHookSetupAgent = hookSetupAgent{
		configKey: config.HookAgentCodexKey,
	}
	claudeCodeHookSetupAgent = hookSetupAgent{
		configKey: config.HookAgentClaudeCodeKey,
	}
)

func updateHookSetupConfigForDefaultPath(agent hookSetupAgent, opts hookSetupOptions) error {
	path, err := config.ConfigFilePath()
	if err != nil {
		return err
	}
	return updateHookSetupConfig(path, agent, opts)
}

func updateHookSetupConfig(configPath string, agent hookSetupAgent, opts hookSetupOptions) error {
	root, err := config.ReadConfigMap(configPath)
	if err != nil {
		return err
	}

	hooks := config.StringMapValue(root[config.HooksKey])
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	agentCfg := config.StringMapValue(hooks[agent.configKey])
	if agentCfg == nil {
		agentCfg = make(map[string]interface{})
	}
	agentCfg[config.HookSetupKey] = true

	if opts.ignoredEventsChanged {
		events := parseIgnoredEvents(opts.ignoredEvents)
		if len(events) == 0 {
			delete(agentCfg, config.HookIgnoredEventsKey)
		} else {
			agentCfg[config.HookIgnoredEventsKey] = events
		}
	}

	hooks[agent.configKey] = agentCfg
	root[config.HooksKey] = hooks

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	return config.WriteConfigMap(configPath, root)
}

func parseIgnoredEvents(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	events := make([]string, 0, len(parts))
	for _, part := range parts {
		event := strings.TrimSpace(part)
		if event == "" {
			continue
		}
		events = append(events, event)
	}
	if len(events) == 0 {
		return nil
	}
	return events
}
