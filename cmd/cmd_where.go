package cmd

import (
	"fmt"
	"os"

	"github.com/lba-studio/n-cli/cmd/where"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewWhereCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "where",
		Short: "Prints out where everything related to n-cli is.",
		Run: func(cmd *cobra.Command, args []string) {
			executable, _ := os.Executable()
			fmt.Printf("Executable: %s\n", executable)
			fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
		},
	}
	c.AddCommand(where.NewWhereConfigCmd())
	return c
}
