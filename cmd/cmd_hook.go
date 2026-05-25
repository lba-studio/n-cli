package cmd

import (
	"github.com/lba-studio/n-cli/cmd/hook"
	"github.com/spf13/cobra"
)

func NewHookCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "hook",
		Short: "Process LLM agent hook events from stdin and send notifications.",
	}
	c.AddCommand(hook.NewHookCursorCmd())
	c.AddCommand(hook.NewHookCodexCmd())
	c.AddCommand(hook.NewHookClaudeCodeCmd())
	return c
}
