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
	maxVersion     = 1 // ensures that hooks.json version changes do not break our integration
)

var hookScriptContent = `#!/bin/sh
# n-cli Cursor hook: notifies via n-cli send --stdin when agent stops or session ends.
# Receives JSON on stdin from Cursor (hook_event_name, status for stop, etc.).
input=""
while IFS= read -r line || [ -n "$line" ]; do
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
    msg="Cursor hook: ${event:-unknown}\nFull payload: ${input:-no_input}"
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
	root := make(map[string]interface{})

	data, err := os.ReadFile(hooksJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse existing %s: %w", hooksJSONPath, err)
		}
	}

	// check version for compatibility with Cursor's hooks.json format
	version, hasVersion := root["version"]
	versionFloat, versionFloatOk := version.(float64)
	versionInt := 0
	if versionFloatOk {
		version = int(versionFloat)
	}
	if !hasVersion {
		root["version"] = 1
	} else if versionInt > maxVersion {
		fmt.Printf("versionInt=%d version=%f\n", versionInt, version)
		// prompt user if they want to proceed anyways
		overwrite, err := prompt.New().
			Ask(fmt.Sprintf("Hooks.json has a newer version %v, which is not yet officially supported (please do log a GitHub Issue in our repo). Proceed anyways?", version)).
			Choose([]string{"Yes", "No"})
		if err != nil {
			return fmt.Errorf("prompt: %w", err)
		}
		if overwrite != "Yes" {
			return fmt.Errorf("hooks.json version %d is not supported", version)
		}
	}

	// process hooks (either merge it or init it)
	hooks, _ := root["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
		root["hooks"] = hooks
	}

	// merge the hooks for each event
	for _, event := range []string{"stop", "sessionEnd"} {
		merged := mergeHookEntry(hooks[event], hookCommand)
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

func mergeHookEntry(existing interface{}, command string) []interface{} {
	entry := map[string]string{"command": command}
	switch existingObject := existing.(type) {
	case []interface{}:
		for _, i := range existingObject {
			if m, ok := i.(map[string]interface{}); ok {
				if c, _ := m["command"].(string); c == command {
					// then it exists, so we return the existing object
					return existingObject
				}
			}
		}
		return append(existingObject, entry)
	}
	return []interface{}{entry}
}

func printSuccess(hooksJSONPath, scriptPath string) {
	fmt.Printf("Cursor hooks configured.\n")
	fmt.Printf("  %s\n", hooksJSONPath)
	fmt.Printf("  %s\n", scriptPath)
	fmt.Println("Restart Cursor for hooks to take effect. Ensure n-cli stays on PATH in the shell Cursor uses.")
}
