package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type codexHookPayload struct {
	HookEventName string `json:"hook_event_name"`
	ToolName      string `json:"tool_name"`
}

func NewHookCodexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codex",
		Short: "Process a Codex hook event from stdin and send a notification.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				fmt.Fprintf(os.Stderr, "n-cli hook error: read stdin: %s\n", err)
				return
			}
			if err := HandleHookCodex(data); err != nil {
				fmt.Fprintf(os.Stderr, "n-cli hook error: %s\n", err)
			}
		},
	}
}

func HandleHookCodex(data []byte) error {
	var payload codexHookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse hook JSON: %w", err)
	}

	msg := FormatCodexMessage(payload)
	if msg == "" {
		return nil
	}

	return notify(msg)
}

func FormatCodexMessage(payload codexHookPayload) string {
	switch payload.HookEventName {
	case "PermissionRequest":
		tool := payload.ToolName
		if tool == "" {
			tool = "unknown tool"
		}
		return fmt.Sprintf("Codex needs approval for: %s", tool)
	case "Stop":
		return "Codex agent finished"
	default:
		event := payload.HookEventName
		if event == "" {
			event = "unknown"
		}
		return fmt.Sprintf("Codex hook: %s", event)
	}
}
