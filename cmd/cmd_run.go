package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/lba-studio/n-cli/pkg/notifier/marker"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Runs arbitrary shell commands through n-cli.",
		Run: func(cobraCmd *cobra.Command, args []string) {
			command := args[0]
			commandArgs := args[1:]

			cmd := exec.Command(command, commandArgs...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			m := marker.NewNotificationMarker(cmd)
			defer m.Done()

			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				os.Exit(1)
			}
		},
	}
}