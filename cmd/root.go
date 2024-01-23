package cmd

import (
	"fmt"
	"os"

	"github.com/lba-studio/n-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "n-cli",
	Short: "Send messages to yourself.",
	Long:  "N is a utility CLI that allows you to send yourself a push notification with any arbitrary message.",

	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Encountered err while displaying --help: %s\n", err.Error())
		}
		os.Exit(1)
	},
}

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.n-cli/config.yaml)")
	rootCmd.AddCommand(NewSendCmd())
	rootCmd.AddCommand(NewWhereCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("$HOME/.n-cli")
		viper.SetConfigName("config")
		if err := config.InitConfigWhenMissing(); err != nil {
			fmt.Println("Exiting - config is missing and we cannot initialise a new one.")
			os.Exit(1)
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
		fmt.Printf("WARN: Cannot read config: %s\n", err.Error())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
