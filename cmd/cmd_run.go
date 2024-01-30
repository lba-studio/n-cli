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
		Args:    cobra.MinimumNArgs(1),
		Short:   "Runs arbitrary shell commands through n-cli.",
		Long:    "Runs arbitrary shell commands. For example, `n-cli run echo Hello, world!` If you'd like to pass in flags to your shell command, use `n-cli run mycommand -- --flag1=true --flag2`.",
		Run: func(cobraCmd *cobra.Command, args []string) {
			command := args[0]
			commandArgs := args[1:]

			cmd := exec.Command(command, commandArgs...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			defer func() {
				os.Exit(cmd.ProcessState.ExitCode())
			}()

			m := marker.NewNotificationMarker(cmd)
			defer m.Done()

			err := cmd.Run()
			if err != nil {
				fmt.Printf("n-cli run error: %s\n", err.Error())
			}
		},
	}
}
