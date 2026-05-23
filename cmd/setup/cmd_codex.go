package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cqroot/prompt"
	"github.com/spf13/cobra"
)

const (
	codexDir = ".codex"
)

var codexHookEvents = []string{"Stop", "PermissionRequest"}

var codexHookScriptContent = `#!/bin/sh
# n-cli Codex hook: notifies via n-cli send --stdin on PermissionRequest / Stop.
# Receives JSON on stdin from Codex (hook_event_name, tool_name, etc.).
input=""
while IFS= read -r line || [ -n "$line" ]; do
  input="${input}${line}"
done

event=$(echo "$input" | sed -n 's/.*"hook_event_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
tool=$(echo "$input" | sed -n 's/.*"tool_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')

case "$event" in
  PermissionRequest)
    msg="Codex needs approval for: ${tool:-unknown tool}"
    ;;
  Stop)
    msg="Codex agent finished"
    ;;
  *)
    msg="Codex hook: ${event:-unknown}"
    ;;
esac

echo "$msg" | n-cli send --stdin 2>/dev/null || true
exit 0
`

func NewSetupCodexCmd() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "codex",
		Short: "Set up Codex hooks so you get n-cli notifications when the agent finishes or needs approval.",
		Long: `Writes user-level Codex hooks to ~/.codex/ so that when the agent loop ends (Stop)
or needs your approval (PermissionRequest), a script runs and calls n-cli send --stdin to notify you.

Requires n-cli to be on PATH. After setup, keep n-cli on PATH in the
shell environment Codex uses so hooks can run.

After setup, review and trust the hooks in Codex with /hooks if prompted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupCodex(force)
		},
	}
	c.Flags().BoolVar(&force, "force", false, "Overwrite existing hook script without prompting")
	return c
}

func runSetupCodex(force bool) error {
	if _, err := exec.LookPath("n-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "n-cli is not on PATH; add it to your PATH so Codex can run it from hooks.\n")
		fmt.Fprintf(os.Stderr, "Check your shell profile (e.g. ~/.zshrc) and restart Codex after setup.\n")
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory: %w", err)
	}
	codexPath := filepath.Join(home, codexDir)
	hooksPath := filepath.Join(codexPath, cursorHooksDir)
	hooksJSONPath := filepath.Join(codexPath, hooksJSON)
	scriptPath := filepath.Join(hooksPath, hookScriptName)
	hookCommand := scriptPath

	if err := os.MkdirAll(hooksPath, 0700); err != nil {
		return fmt.Errorf("create %s: %w", hooksPath, err)
	}

	if err := mergeCodexHooksJSON(hooksJSONPath, hookCommand); err != nil {
		return err
	}

	if _, err := os.Stat(scriptPath); err == nil && !force {
		overwrite, err := prompt.New().
			Ask(fmt.Sprintf("Hook script already exists at %s. Overwrite?", scriptPath)).
			Choose([]string{"Yes", "No"})
		if err != nil {
			return fmt.Errorf("prompt: %w", err)
		}
		if overwrite != "Yes" {
			fmt.Println("Skipping hook script. Run with --force to overwrite.")
			printCodexSuccess(hooksJSONPath, scriptPath)
			return nil
		}
	}

	if err := os.WriteFile(scriptPath, []byte(codexHookScriptContent), 0700); err != nil {
		return fmt.Errorf("write hook script: %w", err)
	}
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("chmod hook script: %w", err)
	}

	printCodexSuccess(hooksJSONPath, scriptPath)
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

	for _, g := range groups {
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
				return groups
			}
		}
	}

	return append(groups, newGroup)
}

func printCodexSuccess(hooksJSONPath, scriptPath string) {
	fmt.Printf("Codex hooks configured.\n")
	fmt.Printf("  %s\n", hooksJSONPath)
	fmt.Printf("  %s\n", scriptPath)
	fmt.Println("Restart Codex for hooks to take effect. Review and trust hooks with /hooks if prompted.")
	fmt.Println("Ensure n-cli stays on PATH in the shell Codex uses.")
}
