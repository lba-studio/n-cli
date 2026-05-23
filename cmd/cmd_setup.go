package cmd

import (
	"github.com/lba-studio/n-cli/cmd/setup"
	"github.com/spf13/cobra"
)

func NewSetupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "setup",
		Short: "Set up integrations (e.g. Cursor, Codex, and Claude Code hooks).",
	}
	c.AddCommand(setup.NewSetupCursorCmd())
	c.AddCommand(setup.NewSetupCodexCmd())
	c.AddCommand(setup.NewSetupClaudeCodeCmd())
	return c
}
