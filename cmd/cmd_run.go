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
		Short:   "Run shell commands and get notified once it's done.",
		Long: `Runs arbitrary shell commands. Once it finishes, n-cli will send a notification to you. It will also capture useful metrics such as the time it took for your command to complete.

Example: n-cli run echo Hello, world!

If you'd like to pass in flags to your shell command, use n-cli run mycommand -- --flag1=true --flag2.

Do note that some metrics may be missing on Windows (open a PR if you are interested in implementing them).
`,
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
