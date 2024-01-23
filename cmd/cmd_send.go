package cmd

import (
	"fmt"
	"strings"

	"github.com/lba-studio/n-cli/pkg/notifier"
	"github.com/spf13/cobra"
)

func NewSendCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "send",
		Short:   "Sends your notification.",
		Aliases: []string{"s"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			msg := strings.Join(args, " ")
			if err := notifier.Notify(msg); err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
				fmt.Print("\n---\nSend failed. Please check the error message(s) above, and verify that your configuration file is correct. \n---\n\n")
			}
		},
	}
}
