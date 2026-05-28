package hook

import (
	"encoding/json"
	"fmt"
	"io"

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
