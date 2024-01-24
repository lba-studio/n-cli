package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cqroot/prompt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type DiscordConfig struct {
	WebhookURL    string `mapstructure:"webhookUrl"`
	MessageFormat string `mapstructure:"messageFormat"`
}

// Config struct to hold the configuration values
type Config struct {
	Discord *DiscordConfig `mapstructure:"discord"`
}

func GetConfig() (cfg Config, err error) {
	err = viper.Unmarshal(&cfg)
	return
}

func onInitFail(tag string, err error) {
	fmt.Printf("%s Cannot init new config: %s\n", tag, err.Error())
}

func InitConfigWhenMissing() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		onInitFail("check.homedirectory", err)
		return err
	}
	// Define the file path
	dirPath := filepath.Join(homeDir, ".n-cli")
	filePath := filepath.Join(dirPath, "config.yaml")

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return nil
	}
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		onInitFail("init.mkdir", err)
	}

	cfg := Config{}

	// fill in the config
	useDiscord, err := prompt.New().
		Ask("Do you want to use Discord?").
		Choose([]string{"Yes", "No"})
	if err != nil {
		onInitFail("prompt.usediscord", err)
		return err
	}

	if useDiscord == "Yes" {
		discordCfg, err := initDiscordConfig()
		if err != nil {
			onInitFail("init.discord", err)
			return err
		}
		cfg.Discord = discordCfg
	}

	var cfgMap map[string]interface{}
	if err := mapstructure.Decode(&cfg, &cfgMap); err != nil {
		onInitFail("marshal.mapstructure", err)
		return err
	}

	// marshal
	cfgBytes, err := yaml.Marshal(&cfgMap)
	if err != nil {
		onInitFail("marshal.yaml", err)
		return err
	}

	if err := os.WriteFile(filePath, cfgBytes, 0644); err != nil {
		onInitFail("writefile", err)
		return err
	}
	return nil
}

func initDiscordConfig() (*DiscordConfig, error) {
	webhookURL, err := prompt.New().Ask("What is your webhook URL?").Input("")
	if err != nil {
		return nil, err
	}
	return &DiscordConfig{
		WebhookURL: webhookURL,
	}, nil
}

// func prompt(prompt string) (string, error) {
// 	fmt.Print(prompt + " ")
// 	scanner := bufio.NewScanner(os.Stdin)
// 	if scanner.Scan() {
// 		return scanner.Text(), nil
// 	} else {
// 		fmt.Println("Error reading input:", scanner.Err())
// 		return "", scanner.Err()
// 	}
// }
