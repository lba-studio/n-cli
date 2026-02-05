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
	cursorDir      = ".cursor"
	cursorHooksDir = "hooks"
	hooksJSON      = "hooks.json"
	hookScriptName = "n-cli-notify.sh"
	hookCommand    = "./hooks/n-cli-notify.sh"
)

var hookScriptContent = `#!/bin/sh
# n-cli Cursor hook: notifies via n-cli send --stdin when agent stops or session ends.
# Receives JSON on stdin from Cursor (hook_event_name, status for stop, etc.).
input=""
while IFS= read -r line; do
  input="${input}${line}"
done

event=$(echo "$input" | sed -n 's/.*"hook_event_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
status=$(echo "$input" | sed -n 's/.*"status"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')

case "$event" in
  stop)
    msg="Done: agent finished (status: ${status:-unknown})"
    ;;
  sessionEnd)
    msg="Session ended"
    ;;
  *)
    msg="Cursor hook: ${event:-unknown}"
    ;;
esac

echo "$msg" | n-cli send --stdin 2>/dev/null || true
exit 0
`

func NewSetupCursorCmd() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "cursor",
		Short: "Set up Cursor hooks so you get n-cli notifications when the agent finishes or the session ends.",
		Long: `Writes user-level Cursor hooks to ~/.cursor/ so that when the agent loop ends (stop)
or the session ends (sessionEnd), a script runs and calls n-cli send --stdin to notify you.

Requires n-cli to be on PATH. After setup, keep n-cli on PATH in the
shell environment Cursor uses so hooks can run.

To get notifications when the agent needs your input, add a per-project rule in .cursor/rules/
that tells the agent to run: echo "Need input: <question>" | n-cli send --stdin`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupCursor(force)
		},
	}
	c.Flags().BoolVar(&force, "force", false, "Overwrite existing hook script without prompting")
	return c
}

func runSetupCursor(force bool) error {
	if _, err := exec.LookPath("n-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "n-cli is not on PATH; add it to your PATH so Cursor can run it from hooks.\n")
		fmt.Fprintf(os.Stderr, "Check your shell profile (e.g. ~/.zshrc) and restart Cursor after setup.\n")
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory: %w", err)
	}
	cursorPath := filepath.Join(home, cursorDir)
	hooksPath := filepath.Join(cursorPath, cursorHooksDir)
	hooksJSONPath := filepath.Join(cursorPath, hooksJSON)
	scriptPath := filepath.Join(hooksPath, hookScriptName)

	if err := os.MkdirAll(hooksPath, 0700); err != nil {
		return fmt.Errorf("create %s: %w", hooksPath, err)
	}

	if err := mergeHooksJSON(hooksJSONPath); err != nil {
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
			printSuccess(hooksJSONPath, scriptPath)
			return nil
		}
	}

	if err := os.WriteFile(scriptPath, []byte(hookScriptContent), 0700); err != nil {
		return fmt.Errorf("write hook script: %w", err)
	}
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("chmod hook script: %w", err)
	}

	printSuccess(hooksJSONPath, scriptPath)
	return nil
}

func mergeHooksJSON(hooksJSONPath string) error {
	var root struct {
		Version int                    `json:"version"`
		Hooks   map[string]interface{} `json:"hooks"`
	}
	root.Version = 1
	root.Hooks = make(map[string]interface{})

	data, err := os.ReadFile(hooksJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse existing %s: %w", hooksJSONPath, err)
		}
		if root.Hooks == nil {
			root.Hooks = make(map[string]interface{})
		}
	}

	for _, event := range []string{"stop", "sessionEnd"} {
		root.Hooks[event] = mergeHookEntry(root.Hooks[event], hookCommand)
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

func mergeHookEntry(existing interface{}, command string) []map[string]string {
	entry := map[string]string{"command": command}
	var list []map[string]string
	switch v := existing.(type) {
	case []interface{}:
		for _, i := range v {
			if m, ok := i.(map[string]interface{}); ok {
				if c, _ := m["command"].(string); c == command {
					return sliceFromInterface(v)
				}
				list = append(list, mapFromInterface(m))
			}
		}
	case nil:
	default:
		if arr, ok := existing.([]map[string]string); ok {
			for _, m := range arr {
				if m["command"] == command {
					return arr
				}
				list = append(list, m)
			}
		}
	}
	list = append(list, entry)
	return list
}

func sliceFromInterface(s []interface{}) []map[string]string {
	out := make([]map[string]string, 0, len(s))
	for _, i := range s {
		if m, ok := i.(map[string]interface{}); ok {
			out = append(out, mapFromInterface(m))
		}
	}
	return out
}

func mapFromInterface(m map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func printSuccess(hooksJSONPath, scriptPath string) {
	fmt.Printf("Cursor hooks configured.\n")
	fmt.Printf("  %s\n", hooksJSONPath)
	fmt.Printf("  %s\n", scriptPath)
	fmt.Println("Restart Cursor for hooks to take effect. Ensure n-cli stays on PATH in the shell Cursor uses.")
}
