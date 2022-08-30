package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	"github.com/medibloc/panacea-doracle/types"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	homeDir string
	rootCmd = &cobra.Command{
		Use:   "doracled",
		Short: "doracle daemon",
	}

	nodePrivKeyPath   string
	oraclePrivKeyPath string
)

func Execute() error {
	return rootCmd.Execute()
}

func initLogger(conf *config.Config) error {
	logLevel, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	return nil
}

// init is run automatically when the package is loaded.
func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultAppHomeDir := filepath.Join(userHomeDir, ".doracle")

	nodePrivKeyPath = filepath.Join(homeDir, types.DefaultNodePrivKeyName)
	oraclePrivKeyPath = filepath.Join(homeDir, types.DefaultOraclePrivKeyName)

	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultAppHomeDir, "application home directory")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(genOracleKeyCmd)
	rootCmd.AddCommand(verifyReport)
	rootCmd.AddCommand(registerOracleCmd())
	rootCmd.AddCommand(getOracleKeyCmd)
}
