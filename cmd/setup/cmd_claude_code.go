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
	claudeCodeDir         = ".claude"
	claudeCodeSettings    = "settings.json"
	claudeCodeStopEvent   = "Stop"
	claudeCodeNotifyEvent = "Notification"
)

var claudeCodeHookScriptContent = `#!/bin/sh
# n-cli Claude Code hook: notifies via n-cli send --stdin on Notification / Stop.
# Receives JSON on stdin from Claude Code (hook_event_name, notification_type, etc.).
input=""
while IFS= read -r line || [ -n "$line" ]; do
  input="${input}${line}"
done

event=$(echo "$input" | sed -n 's/.*"hook_event_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
notification_type=$(echo "$input" | sed -n 's/.*"notification_type"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
message=$(echo "$input" | sed -n 's/.*"message"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')

case "$event" in
  Notification)
    case "$notification_type" in
      permission_prompt)
        msg="Claude Code needs approval: ${message:-permission needed}"
        ;;
      *)
        msg="Claude Code: ${message:-notification}"
        ;;
    esac
    ;;
  Stop)
    msg="Claude Code agent finished"
    ;;
  *)
    msg="Claude Code hook: ${event:-unknown}"
    ;;
esac

echo "$msg" | n-cli send --stdin 2>/dev/null || true
exit 0
`

func NewSetupClaudeCodeCmd() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "claude-code",
		Short: "Set up Claude Code hooks so you get n-cli notifications when the agent finishes or needs approval.",
		Long: `Writes user-level Claude Code hooks to ~/.claude/ so that when the agent loop ends (Stop)
or needs your approval (Notification with permission_prompt), a script runs and calls n-cli send --stdin to notify you.

Requires n-cli to be on PATH. After setup, keep n-cli on PATH in the
shell environment Claude Code uses so hooks can run.

After setup, review and trust the hooks in Claude Code with /hooks if prompted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupClaudeCode(force)
		},
	}
	c.Flags().BoolVar(&force, "force", false, "Overwrite existing hook script without prompting")
	return c
}

func runSetupClaudeCode(force bool) error {
	if _, err := exec.LookPath("n-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "n-cli is not on PATH; add it to your PATH so Claude Code can run it from hooks.\n")
		fmt.Fprintf(os.Stderr, "Check your shell profile (e.g. ~/.zshrc) and restart Claude Code after setup.\n")
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory: %w", err)
	}
	claudePath := filepath.Join(home, claudeCodeDir)
	hooksPath := filepath.Join(claudePath, cursorHooksDir)
	settingsPath := filepath.Join(claudePath, claudeCodeSettings)
	scriptPath := filepath.Join(hooksPath, hookScriptName)
	hookCommand := scriptPath

	if err := os.MkdirAll(hooksPath, 0700); err != nil {
		return fmt.Errorf("create %s: %w", hooksPath, err)
	}

	if err := mergeClaudeCodeSettings(settingsPath, hookCommand); err != nil {
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
			printClaudeCodeSuccess(settingsPath, scriptPath)
			return nil
		}
	}

	if err := os.WriteFile(scriptPath, []byte(claudeCodeHookScriptContent), 0700); err != nil {
		return fmt.Errorf("write hook script: %w", err)
	}
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("chmod hook script: %w", err)
	}

	printClaudeCodeSuccess(settingsPath, scriptPath)
	return nil
}

func mergeClaudeCodeSettings(settingsPath, hookCommand string) error {
	root := make(map[string]interface{})

	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse existing %s: %w", settingsPath, err)
		}
	}

	hooks, _ := root["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
		root["hooks"] = hooks
	}

	hooks[claudeCodeStopEvent] = mergeClaudeCodeHookEvent(hooks[claudeCodeStopEvent], hookCommand, "")
	hooks[claudeCodeNotifyEvent] = mergeClaudeCodeHookEvent(hooks[claudeCodeNotifyEvent], hookCommand, "permission_prompt")

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.json: %w", err)
	}
	if err := os.WriteFile(settingsPath, out, 0644); err != nil {
		return fmt.Errorf("write %s: %w", settingsPath, err)
	}
	return nil
}

func mergeClaudeCodeHookEvent(existing interface{}, command, matcher string) []interface{} {
	newHandler := map[string]interface{}{
		"type":    "command",
		"command": command,
	}
	newGroup := map[string]interface{}{
		"hooks": []interface{}{newHandler},
	}
	if matcher != "" {
		newGroup["matcher"] = matcher
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
		if !claudeCodeMatcherMatches(group["matcher"], matcher) {
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

func claudeCodeMatcherMatches(existingMatcher interface{}, expectedMatcher string) bool {
	existing, _ := existingMatcher.(string)
	if expectedMatcher == "" {
		return existing == ""
	}
	return existing == expectedMatcher
}

func printClaudeCodeSuccess(settingsPath, scriptPath string) {
	fmt.Printf("Claude Code hooks configured.\n")
	fmt.Printf("  %s\n", settingsPath)
	fmt.Printf("  %s\n", scriptPath)
	fmt.Println("Restart Claude Code for hooks to take effect. Review and trust hooks with /hooks if prompted.")
	fmt.Println("Ensure n-cli stays on PATH in the shell Claude Code uses.")
}
