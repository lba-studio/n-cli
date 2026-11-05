package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ConfigFilePath returns the config file in use, falling back to ~/.n-cli/config.yaml.
func ConfigFilePath() (string, error) {
	if path := viper.ConfigFileUsed(); path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".n-cli", "config.yaml"), nil
}

// ReadConfigMap reads the YAML config as a generic map so unknown keys survive a rewrite.
// A missing or empty file yields an empty map.
func ReadConfigMap(configPath string) (map[string]interface{}, error) {
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

// WriteConfigMap writes root as YAML, keeping the existing file's permissions.
func WriteConfigMap(configPath string, root map[string]interface{}) error {
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

// StringMapValue converts a decoded YAML mapping into map[string]interface{}, or returns nil.
func StringMapValue(value interface{}) map[string]interface{} {
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
