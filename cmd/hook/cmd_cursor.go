package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type cursorHookPayload struct {
	HookEventName string `json:"hook_event_name"`
	Status        string `json:"status"`
}

func NewHookCursorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cursor",
		Short: "Process a Cursor hook event from stdin and send a notification.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				fmt.Fprintf(os.Stderr, "n-cli hook error: read stdin: %s\n", err)
				return
			}
			if err := HandleHookCursor(data); err != nil {
				fmt.Fprintf(os.Stderr, "n-cli hook error: %s\n", err)
			}
		},
	}
}

func HandleHookCursor(data []byte) error {
	var payload cursorHookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse hook JSON: %w", err)
	}

	msg := FormatCursorMessage(payload)
	if msg == "" {
		return nil
	}

	return notify(msg)
}

func FormatCursorMessage(payload cursorHookPayload) string {
	switch payload.HookEventName {
	case "stop":
		status := payload.Status
		if status == "" {
			status = "unknown"
		}
		return fmt.Sprintf("Done: agent finished (status: %s)", status)
	case "sessionEnd":
		return ""
	default:
		event := payload.HookEventName
		if event == "" {
			event = "unknown"
		}
		return fmt.Sprintf("Cursor hook: %s", event)
	}
}
