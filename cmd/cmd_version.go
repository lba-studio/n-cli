package cmd

import (
	"fmt"

	"github.com/lba-studio/n-cli/pkg/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Get n-cli version.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersion())
		},
	}
}
