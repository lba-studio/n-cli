package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type claudeCodeHookPayload struct {
	HookEventName    string `json:"hook_event_name"`
	NotificationType string `json:"notification_type"`
	Message          string `json:"message"`
	// Cursor also invokes hooks registered in ~/.claude/settings.json when a
	// Cursor agent finishes. Those payloads use Cursor conventions (e.g.
	// hook_event_name "stop") and include cursor_version; Claude Code uses
	// "Stop" and does not send this field.
	// @verzac: this is probably a buggy behaviour from Cursor, but let's guard against it anyways. Codex seems safe from this haha
	CursorVersion string `json:"cursor_version"`
}

func NewHookClaudeCodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude-code",
		Short: "Process a Claude Code hook event from stdin and send a notification.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				fmt.Fprintf(os.Stderr, "n-cli hook error: read stdin: %s\n", err)
				return
			}
			if err := HandleHookClaudeCode(data); err != nil {
				fmt.Fprintf(os.Stderr, "n-cli hook error: %s\n", err)
			}
		},
	}
}

func HandleHookClaudeCode(data []byte) error {
	var payload claudeCodeHookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse hook JSON: %w", err)
	}

	// Ignore Cursor-originated payloads to avoid duplicate notifications when
	// both n-cli setup cursor and n-cli setup claude-code are configured.
	if payload.CursorVersion != "" {
		return nil
	}

	ignored, err := isHookEventIgnored(hookAgentClaudeCode, payload.HookEventName)
	if err != nil {
		return err
	}
	if ignored {
		return nil
	}

	msg := FormatClaudeCodeMessage(payload)
	if msg == "" {
		return nil
	}

	return notify(msg)
}

func FormatClaudeCodeMessage(payload claudeCodeHookPayload) string {
	switch payload.HookEventName {
	case "Notification":
		switch payload.NotificationType {
		case "permission_prompt":
			message := payload.Message
			if message == "" {
				message = "permission needed"
			}
			return fmt.Sprintf("Claude Code needs approval: %s", message)
		default:
			message := payload.Message
			if message == "" {
				message = "notification"
			}
			return fmt.Sprintf("Claude Code: %s", message)
		}
	case "Stop":
		return "Claude Code agent finished"
	default:
		event := payload.HookEventName
		if event == "" {
			event = "unknown"
		}
		return fmt.Sprintf("Claude Code hook: %s", event)
	}
}
