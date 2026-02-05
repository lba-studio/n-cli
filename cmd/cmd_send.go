package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lba-studio/n-cli/pkg/notifier"
	"github.com/spf13/cobra"
)

func NewSendCmd() *cobra.Command {
	var fromStdin bool
	c := &cobra.Command{
		Use:     "send",
		Short:   "Sends your notification.",
		Aliases: []string{"s"},
		Args:    cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var msg string
			if fromStdin {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err.Error())
					os.Exit(1)
				}
				msg = strings.TrimRight(string(b), "\n")
			} else {
				if len(args) == 0 {
					fmt.Print("message required: pass arguments or use --stdin\n")
					os.Exit(1)
				}
				msg = strings.Join(args, " ")
			}
			if err := notifier.Notify(msg); err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
				fmt.Print("\n---\nSend failed. Please check the error message(s) above, and verify that your configuration file is correct. \n---\n\n")
			}
		},
	}
	c.Flags().BoolVar(&fromStdin, "stdin", false, "Read message from stdin instead of arguments")
	return c
}
