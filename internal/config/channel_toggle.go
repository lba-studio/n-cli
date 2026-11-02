package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	disabledKey = "disabled"
	systemKey   = "system"
	customKey   = "custom"
	customsKey  = "customs"
)

var customsIndexPattern = regexp.MustCompile(`^customs\[(\d+)\]$`)

// SetChannelDisabled sets (disabled=true) or removes (disabled=false) the `disabled`
// flag of the channel identified by label in the config file at configPath.
//
// Accepted labels: "system", "discord", "slack", "custom" (legacy block),
// "customs[i]", or the name of a custom entry. It returns a human-readable
// description of the resolved channel and whether the file was changed.
func SetChannelDisabled(configPath, label string, disabled bool) (resolved string, changed bool, err error) {
	root, err := ReadConfigMap(configPath)
	if err != nil {
		return "", false, err
	}

	target, resolved, err := resolveChannel(root, strings.TrimSpace(label), disabled)
	if err != nil {
		return "", false, err
	}

	if (target[disabledKey] == true) == disabled {
		return resolved, false, nil
	}
	if disabled {
		target[disabledKey] = true
	} else {
		delete(target, disabledKey)
	}
	return resolved, true, WriteConfigMap(configPath, root)
}

// resolveChannel returns the (mutable, attached to root) config map for label.
func resolveChannel(root map[string]interface{}, label string, disabling bool) (map[string]interface{}, string, error) {
	switch label {
	case "":
		return nil, "", fmt.Errorf("label must not be empty")
	case systemKey:
		// The system notifier is on by default, so disabling it may need a new block.
		m := StringMapValue(root[systemKey])
		if m == nil {
			m = map[string]interface{}{}
			if disabling {
				root[systemKey] = m
			}
		} else {
			root[systemKey] = m
		}
		return m, systemKey, nil
	case "discord", "slack", customKey:
		m := StringMapValue(root[label])
		if m == nil {
			return nil, "", fmt.Errorf("%q is not configured", label)
		}
		root[label] = m
		return m, describe(label, m), nil
	}

	customs, err := customsEntries(root)
	if err != nil {
		return nil, "", err
	}

	if match := customsIndexPattern.FindStringSubmatch(label); match != nil {
		i, err := strconv.Atoi(match[1])
		if err != nil || i >= len(customs) {
			return nil, "", fmt.Errorf("%s does not exist (%d customs entries configured)", label, len(customs))
		}
		return customs[i], describe(label, customs[i]), nil
	}

	type candidate struct {
		label string
		cfg   map[string]interface{}
	}
	var matches []candidate
	if legacy := StringMapValue(root[customKey]); legacy != nil && legacy["name"] == label {
		root[customKey] = legacy
		matches = append(matches, candidate{customKey, legacy})
	}
	for i, entry := range customs {
		if entry["name"] == label {
			matches = append(matches, candidate{fmt.Sprintf("customs[%d]", i), entry})
		}
	}

	switch len(matches) {
	case 0:
		return nil, "", fmt.Errorf("no channel matches %q; use system, discord, slack, custom, customs[i] or a custom entry's name", label)
	case 1:
		return matches[0].cfg, describe(matches[0].label, matches[0].cfg), nil
	default:
		labels := make([]string, len(matches))
		for i, m := range matches {
			labels[i] = m.label
		}
		return nil, "", fmt.Errorf("%q is ambiguous; use one of: %s", label, strings.Join(labels, ", "))
	}
}

// customsEntries normalises root["customs"] in place and returns its entries as maps.
func customsEntries(root map[string]interface{}) ([]map[string]interface{}, error) {
	raw, ok := root[customsKey]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s must be a list", customsKey)
	}
	entries := make([]map[string]interface{}, len(list))
	for i, item := range list {
		m := StringMapValue(item)
		if m == nil {
			return nil, fmt.Errorf("%s[%d] must be a mapping", customsKey, i)
		}
		list[i] = m
		entries[i] = m
	}
	return entries, nil
}

func describe(label string, cfg map[string]interface{}) string {
	if name, ok := cfg["name"].(string); ok && name != "" && name != label {
		return fmt.Sprintf("%s (%s)", label, name)
	}
	return label
}
