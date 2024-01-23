package where

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewWhereConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Prints out where your config is.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", viper.ConfigFileUsed())
		},
	}
}
