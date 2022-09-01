package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/client"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	configFileName = "config.toml"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configs in home dir",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetContext(cmd)
		if err != nil {
			return err
		}

		homeDir := ctx.HomeDir
		if _, err := os.Stat(homeDir); err == nil {
			return fmt.Errorf("home dir(%v) already exists", homeDir)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check home dir: %w", err)
		}

		if err := os.MkdirAll(homeDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create config dir: %w", err)
		}

		if _, err := os.Stat(panacea.DbDir); os.IsNotExist(err) {
			err = os.MkdirAll(panacea.DbDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create db dir: %w", err)
			}
		}

		return config.WriteConfigTOML(getConfigPath(homeDir), config.DefaultConfig())
	},
}

func getConfigPath(homeDir string) string {
	return filepath.Join(homeDir, configFileName)
}
