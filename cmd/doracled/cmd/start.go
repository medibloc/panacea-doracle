package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/spf13/cobra"
	"time"

	log "github.com/sirupsen/logrus"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.ReadConfigTOML(getConfigPath())
		if err != nil {
			return fmt.Errorf("failed to read config from file: %w", err)
		}

		if err := initLogger(conf); err != nil {
			return fmt.Errorf("failed to init logger: %w", err)
		}

		return nil
	},
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
