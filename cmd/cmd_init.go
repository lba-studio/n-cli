package cmd

import (
	"github.com/lba-studio/n-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure n-cli.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			config.InitConfig(config.InitConfigOptions{
				DoNotSkipIfFileExists: true,
			})
		},
	}
}
