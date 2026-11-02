package cmd

import (
	"fmt"

	"github.com/lba-studio/n-cli/internal/config"
	"github.com/spf13/cobra"
)

const channelLabelHelp = `<label> is one of:
  system, discord, slack   a built-in channel
  custom                   the legacy single "custom" block
  customs[i]               the i-th entry (0-based) of the "customs" list
  <name>                   a custom entry by its "name"`

func NewDisableCmd() *cobra.Command {
	return newChannelToggleCmd(true)
}

func NewEnableCmd() *cobra.Command {
	return newChannelToggleCmd(false)
}

func newChannelToggleCmd(disable bool) *cobra.Command {
	verb, past, setting := "enable", "Enabled", "removes `disabled`"
	if disable {
		verb, past, setting = "disable", "Disabled", "sets `disabled: true`"
	}
	return &cobra.Command{
		Use:   verb + " <label>",
		Short: fmt.Sprintf("%s a notification channel in the config.", past[:len(past)-1]),
		Long: fmt.Sprintf("%s a notification channel without losing its settings (%s in the config file).\n\n%s",
			past[:len(past)-1], setting, channelLabelHelp),
		Example: fmt.Sprintf("  n-cli %[1]s slack\n  n-cli %[1]s customs[0]\n  n-cli %[1]s pagerduty", verb),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			path, err := config.ConfigFilePath()
			if err != nil {
				return err
			}
			resolved, changed, err := config.SetChannelDisabled(path, args[0], disable)
			if err != nil {
				return err
			}
			if !changed {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is already %sd in %s\n", resolved, verb, path)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s in %s\n", past, resolved, path)
			return nil
		},
	}
}
