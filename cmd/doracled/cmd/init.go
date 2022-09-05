package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/medibloc/panacea-doracle/config"
	"github.com/spf13/cobra"
)

const (
	configFileName = "config.toml"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configs in home dir",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(homeDir); err == nil {
			return fmt.Errorf("home dir(%v) already exists", homeDir)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check home dir: %w", err)
		}

		if err := os.MkdirAll(homeDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create config dir: %w", err)
		}

		defaultConfig := config.DefaultConfig()

		if _, err := os.Stat(defaultConfig.DataDir); os.IsNotExist(err) {
			err = os.MkdirAll(defaultConfig.DataDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create db dir: %w", err)
			}
		}

		return config.WriteConfigTOML(getConfigPath(), defaultConfig)
	},
}

func getConfigPath() string {
	return filepath.Join(homeDir, configFileName)
}
