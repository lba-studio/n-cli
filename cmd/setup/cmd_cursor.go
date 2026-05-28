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
	cursorDir   = ".cursor"
	hooksJSON   = "hooks.json"
	hookCommand = "n-cli hook cursor"
	maxVersion  = 1 // ensures that hooks.json version changes do not break our integration
)

func NewSetupCursorCmd() *cobra.Command {
	var ignoredEvents string
	c := &cobra.Command{
		Use:   "cursor",
		Short: "Set up Cursor hooks so you get n-cli notifications when the agent finishes or the session ends.",
		Long: `Writes user-level Cursor hooks to ~/.cursor/ so that when the agent loop ends (stop)
or the session ends (sessionEnd), n-cli hook cursor runs to notify you.

Requires n-cli to be on PATH. After setup, keep n-cli on PATH in the
shell environment Cursor uses so hooks can run.

To get notifications when the agent needs your input, add a per-project rule in .cursor/rules/
that tells the agent to run: echo "Need input: <question>" | n-cli send --stdin`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupCursor(hookSetupOptions{
				ignoredEvents:        ignoredEvents,
				ignoredEventsChanged: cmd.Flags().Changed("ignored-events"),
			})
		},
	}
	c.Flags().StringVar(&ignoredEvents, "ignored-events", "", "Comma-separated hook event names to skip notifications for.")
	return c
}

func runSetupCursor(opts hookSetupOptions) error {
	if _, err := exec.LookPath("n-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "n-cli is not on PATH; add it to your PATH so Cursor can run it from hooks.\n")
		fmt.Fprintf(os.Stderr, "%s\n", pathHint())
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home directory: %w", err)
	}
	cursorPath := filepath.Join(home, cursorDir)
	hooksJSONPath := filepath.Join(cursorPath, hooksJSON)

	if err := os.MkdirAll(cursorPath, 0700); err != nil {
		return fmt.Errorf("create %s: %w", cursorPath, err)
	}

	if err := mergeHooksJSON(hooksJSONPath); err != nil {
		return err
	}

	if err := updateHookSetupConfigForDefaultPath(cursorHookSetupAgent, opts); err != nil {
		return err
	}

	printSuccess(hooksJSONPath)
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
		versionInt = int(versionFloat)
	}
	if !hasVersion {
		root["version"] = 1
	} else if versionInt > maxVersion {
		overwrite, err := prompt.New().
			Ask(fmt.Sprintf("Hooks.json has a newer version %d, which is not yet officially supported (please do log a GitHub Issue in our repo). Proceed anyways?", versionInt)).
			Choose([]string{"Yes", "No"})
		if err != nil {
			return fmt.Errorf("prompt: %w", err)
		}
		if overwrite != "Yes" {
			return fmt.Errorf("hooks.json version %d is not supported", versionInt)
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
	existingSlice, ok := existing.([]interface{})
	if !ok {
		return []interface{}{entry}
	}

	filtered := filterOldShellHookEntries(existingSlice)
	for _, i := range filtered {
		if hookEntryCommand(i) == command {
			return filtered
		}
	}
	return append(filtered, entry)
}

func hookEntryCommand(entry interface{}) string {
	switch m := entry.(type) {
	case map[string]interface{}:
		c, _ := m["command"].(string)
		return c
	case map[string]string:
		return m["command"]
	default:
		return ""
	}
}

func filterOldShellHookEntries(entries []interface{}) []interface{} {
	filtered := make([]interface{}, 0, len(entries))
	for _, i := range entries {
		if isOldShellHook(hookEntryCommand(i)) {
			continue
		}
		filtered = append(filtered, i)
	}
	return filtered
}

func printSuccess(hooksJSONPath string) {
	fmt.Printf("Cursor hooks configured.\n")
	fmt.Printf("  %s\n", hooksJSONPath)
	fmt.Println("Restart Cursor for hooks to take effect. Ensure n-cli stays on PATH in the shell Cursor uses.")
}
