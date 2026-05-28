package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Claude Code hooks documentation: https://code.claude.com/docs/en/hooks
const (
	claudeCodeDir         = ".claude"
	claudeCodeSettings    = "settings.json"
	claudeCodeStopEvent   = "Stop"
	claudeCodeNotifyEvent = "Notification"
	claudeCodeHookCommand = "n-cli hook claude-code"
)

func NewSetupClaudeCodeCmd() *cobra.Command {
	var ignoredEvents string
	c := &cobra.Command{
		Use:   "claude-code",
		Short: "Set up Claude Code hooks so you get n-cli notifications when the agent finishes or needs approval.",
		Long: `Writes user-level Claude Code hooks to ~/.claude/ so that when the agent loop ends (Stop)
or needs your approval (Notification with permission_prompt), n-cli hook claude-code runs to notify you.

Requires n-cli to be on PATH. After setup, keep n-cli on PATH in the
shell environment Claude Code uses so hooks can run.

After setup, review and trust the hooks in Claude Code with /hooks if prompted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupClaudeCode(hookSetupOptions{
				ignoredEvents:        ignoredEvents,
				ignoredEventsChanged: cmd.Flags().Changed("ignored-events"),
			})
		},
	}
	c.Flags().StringVar(&ignoredEvents, "ignored-events", "", "Comma-separated hook event names to skip notifications for.")
	return c
}

func runSetupClaudeCode(opts hookSetupOptions) error {
	if _, err := exec.LookPath("n-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "n-cli is not on PATH; add it to your PATH so Claude Code can run it from hooks.\n")
		fmt.Fprintf(os.Stderr, "%s\n", pathHint())
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory: %w", err)
	}
	claudePath := filepath.Join(home, claudeCodeDir)
	settingsPath := filepath.Join(claudePath, claudeCodeSettings)

	if err := os.MkdirAll(claudePath, 0700); err != nil {
		return fmt.Errorf("create %s: %w", claudePath, err)
	}

	if err := mergeClaudeCodeSettings(settingsPath, claudeCodeHookCommand); err != nil {
		return err
	}

	if err := updateHookSetupConfigForDefaultPath(claudeCodeHookSetupAgent, opts); err != nil {
		return err
	}

	printClaudeCodeSuccess(settingsPath)
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

	filteredGroups := make([]interface{}, 0, len(groups))
	for _, g := range groups {
		group, ok := g.(map[string]interface{})
		if !ok {
			filteredGroups = append(filteredGroups, g)
			continue
		}
		if !claudeCodeMatcherMatches(group["matcher"], matcher) {
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
				return filteredGroups
			}
		}
	}

	return append(filteredGroups, newGroup)
}

func claudeCodeMatcherMatches(existingMatcher interface{}, expectedMatcher string) bool {
	existing, _ := existingMatcher.(string)
	if expectedMatcher == "" {
		return existing == ""
	}
	return existing == expectedMatcher
}

func printClaudeCodeSuccess(settingsPath string) {
	fmt.Printf("Claude Code hooks configured.\n")
	fmt.Printf("  %s\n", settingsPath)
	fmt.Println("Restart Claude Code for hooks to take effect. Review and trust hooks with /hooks if prompted.")
	fmt.Println("Ensure n-cli stays on PATH in the shell Claude Code uses.")
}
