package cmd

import (
	"github.com/lba-studio/n-cli/cmd/setup"
	"github.com/spf13/cobra"
)

func NewSetupCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "setup",
		Short: "Set up integrations (e.g. Cursor hooks).",
	}
	c.AddCommand(setup.NewSetupCursorCmd())
	return c
}
