package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "doracled",
		Short: "doracle daemon",
	}
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

	rootCmd.PersistentFlags().String(flags.FlagHome, defaultAppHomeDir, "application home directory")

	rootCmd.AddCommand(
		initCmd,
		startCmd(),
		genOracleKeyCmd,
		verifyReport,
		registerOracleCmd(),
		getOracleKeyCmd,
	)
}

func loadConfigFromHome(cmd *cobra.Command) (*config.Config, error) {
	homeDir, err := cmd.Flags().GetString(flags.FlagHome)
	if err != nil {
		return nil, fmt.Errorf("failed to read a home flag: %w", err)
	}

	conf, err := config.ReadConfigTOML(getConfigPath(homeDir))
	if err != nil {
		return nil, fmt.Errorf("failed to read config from file: %w", err)
	}
	conf.HomeDir = homeDir

	if err := initLogger(conf); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	return conf, nil
}
