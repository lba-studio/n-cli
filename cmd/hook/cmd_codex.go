package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type codexHookPayload struct {
	HookEventName        string `json:"hook_event_name"`
	ToolName             string `json:"tool_name"`
	SessionID            string `json:"session_id"`
	Cwd                  string `json:"cwd"`
	LastAssistantMessage string `json:"last_assistant_message"`
}

func NewHookCodexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codex",
		Short: "Process a Codex hook event from stdin and send a notification.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "n-cli hook error: read stdin: %s\n", err)
				return
			}
			output, err := HandleHookCodex(data)
			if len(output) > 0 {
				if _, writeErr := cmd.OutOrStdout().Write(output); writeErr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "n-cli hook error: write output: %s\n", writeErr)
				}
			}
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "n-cli hook error: %s\n", err)
			}
		},
	}
}

func HandleHookCodex(data []byte) ([]byte, error) {
	var payload codexHookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse hook JSON: %w", err)
	}

	var output []byte
	if payload.HookEventName == "Stop" {
		output = []byte("{\"continue\":true}\n")
	}

	ignored, err := isHookEventIgnored(hookAgentCodex, payload.HookEventName)
	if err != nil {
		return output, err
	}
	if ignored {
		return output, nil
	}

	msg := FormatCodexMessage(payload)
	if msg == "" {
		return output, nil
	}

	return output, notify(msg)
}

// codexThreadName looks up the human-readable session name from ~/.codex/session_index.jsonl.
// The file is append-only; renames add a new line, so the last matching entry wins.
func codexThreadName(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return codexThreadNameFromPath(sessionID, filepath.Join(home, ".codex", "session_index.jsonl"))
}

func codexThreadNameFromPath(sessionID, indexPath string) string {
	if sessionID == "" {
		return ""
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return ""
	}
	var name string
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		var entry struct {
			ID         string `json:"id"`
			ThreadName string `json:"thread_name"`
		}
		if json.Unmarshal([]byte(line), &entry) == nil && entry.ID == sessionID {
			name = entry.ThreadName
		}
	}
	return name
}

// codexAgentLabel extracts the project name from the working directory as a fallback label.
func codexAgentLabel(cwd string) string {
	if cwd == "" {
		return ""
	}
	name := filepath.Base(cwd)
	if name == "." || name == "/" || name == "\\" {
		return ""
	}
	return name
}

func FormatCodexMessage(payload codexHookPayload) string {
	label := codexThreadName(payload.SessionID)
	if label == "" {
		label = codexAgentLabel(payload.Cwd)
	}
	prefix := "Codex"
	if label != "" {
		prefix = fmt.Sprintf("Codex [%s]", label)
	}
	switch payload.HookEventName {
	case "PermissionRequest":
		tool := payload.ToolName
		if tool == "" {
			tool = "unknown tool"
		}
		return fmt.Sprintf("%s needs approval for: %s", prefix, tool)
	case "Stop":
		return fmt.Sprintf("%s agent finished", prefix)
	default:
		event := payload.HookEventName
		if event == "" {
			event = "unknown"
		}
		return fmt.Sprintf("%s hook: %s", prefix, event)
	}
}
