package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Codex hooks documentation: https://developers.openai.com/codex/hooks
const (
	codexDir          = ".codex"
	codexHookCommand  = "n-cli hook codex"
)

var codexHookEvents = []string{"Stop", "PermissionRequest"}

func NewSetupCodexCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "codex",
		Short: "Set up Codex hooks so you get n-cli notifications when the agent finishes or needs approval.",
		Long: `Writes user-level Codex hooks to ~/.codex/ so that when the agent loop ends (Stop)
or needs your approval (PermissionRequest), n-cli hook codex runs to notify you.

Requires n-cli to be on PATH. After setup, keep n-cli on PATH in the
shell environment Codex uses so hooks can run.

After setup, review and trust the hooks in Codex with /hooks if prompted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupCodex()
		},
	}
	return c
}

func runSetupCodex() error {
	if _, err := exec.LookPath("n-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "n-cli is not on PATH; add it to your PATH so Codex can run it from hooks.\n")
		fmt.Fprintf(os.Stderr, "%s\n", pathHint())
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory: %w", err)
	}
	codexPath := filepath.Join(home, codexDir)
	hooksJSONPath := filepath.Join(codexPath, hooksJSON)

	if err := os.MkdirAll(codexPath, 0700); err != nil {
		return fmt.Errorf("create %s: %w", codexPath, err)
	}

	if err := mergeCodexHooksJSON(hooksJSONPath, codexHookCommand); err != nil {
		return err
	}

	printCodexSuccess(hooksJSONPath)
	return nil
}

func mergeCodexHooksJSON(hooksJSONPath, hookCommand string) error {
	root := make(map[string]interface{})

	data, err := os.ReadFile(hooksJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse existing %s: %w", hooksJSONPath, err)
		}
	}

	hooks, _ := root["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
		root["hooks"] = hooks
	}

	for _, event := range codexHookEvents {
		merged := mergeCodexHookEvent(hooks[event], hookCommand)
		hooks[event] = merged
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal hooks.json: %w", err)
	}
	if err := os.WriteFile(hooksJSONPath, out, 0644); err != nil {
		return fmt.Errorf("write %s: %w", hooksJSONPath, err)
	}
	return nil
}

func mergeCodexHookEvent(existing interface{}, command string) []interface{} {
	newHandler := map[string]interface{}{
		"type":    "command",
		"command": command,
	}
	newGroup := map[string]interface{}{
		"hooks": []interface{}{newHandler},
	}

	groups, ok := existing.([]interface{})
	if !ok || groups == nil {
		return []interface{}{newGroup}
	}

	filteredGroups := make([]interface{}, 0, len(groups))
	for _, g := range groups {
		group, ok := g.(map[string]interface{})
		if !ok {
			filteredGroups = append(filteredGroups, g)
			continue
		}
		hooksList, ok := group["hooks"].([]interface{})
		if !ok {
			filteredGroups = append(filteredGroups, g)
			continue
		}
		filteredHooks := make([]interface{}, 0, len(hooksList))
		for _, h := range hooksList {
			hook, ok := h.(map[string]interface{})
			if !ok {
				filteredHooks = append(filteredHooks, h)
				continue
			}
			if c, _ := hook["command"].(string); isOldShellHook(c) {
				continue
			}
			filteredHooks = append(filteredHooks, h)
		}
		if len(filteredHooks) == 0 {
			continue
		}
		group["hooks"] = filteredHooks
		filteredGroups = append(filteredGroups, group)
	}

	for _, g := range filteredGroups {
		group, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		hooksList, ok := group["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, h := range hooksList {
			hook, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if c, _ := hook["command"].(string); c == command {
				return filteredGroups
			}
		}
	}

	return append(filteredGroups, newGroup)
}

func printCodexSuccess(hooksJSONPath string) {
	fmt.Printf("Codex hooks configured.\n")
	fmt.Printf("  %s\n", hooksJSONPath)
	fmt.Println("Restart Codex for hooks to take effect. Review and trust hooks with /hooks if prompted.")
	fmt.Println("Ensure n-cli stays on PATH in the shell Codex uses.")
}
