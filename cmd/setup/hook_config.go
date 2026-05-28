package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lba-studio/n-cli/internal/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

func hookConfigPath() (string, error) {
	if path := viper.ConfigFileUsed(); path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".n-cli", "config.yaml"), nil
}

func updateHookSetupConfigForDefaultPath(agent hookSetupAgent, opts hookSetupOptions) error {
	path, err := hookConfigPath()
	if err != nil {
		return err
	}
	return updateHookSetupConfig(path, agent, opts)
}

func updateHookSetupConfig(configPath string, agent hookSetupAgent, opts hookSetupOptions) error {
	root, err := readConfigMap(configPath)
	if err != nil {
		return err
	}

	hooks := stringMapValue(root[config.HooksKey])
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	agentCfg := stringMapValue(hooks[agent.configKey])
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
	return writeConfigMap(configPath, root)
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

func readConfigMap(configPath string) (map[string]interface{}, error) {
	root := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return root, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", configPath, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return root, nil
	}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse %s: %w", configPath, err)
	}
	if root == nil {
		root = make(map[string]interface{})
	}
	return root, nil
}

func writeConfigMap(configPath string, root map[string]interface{}) error {
	perm := os.FileMode(0644)
	if info, err := os.Stat(configPath); err == nil {
		perm = info.Mode().Perm()
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, out, perm); err != nil {
		return fmt.Errorf("write %s: %w", configPath, err)
	}
	return nil
}

func stringMapValue(value interface{}) map[string]interface{} {
	switch m := value.(type) {
	case map[string]interface{}:
		return m
	case map[interface{}]interface{}:
		converted := make(map[string]interface{}, len(m))
		for k, v := range m {
			key, ok := k.(string)
			if !ok {
				continue
			}
			converted[key] = v
		}
		return converted
	default:
		return nil
	}
}
