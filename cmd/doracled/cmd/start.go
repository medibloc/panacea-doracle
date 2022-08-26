package cmd

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
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

		subscriber, err := event.NewSubscriber(conf)
		if err != nil {
			return err
		}
		defer subscriber.Close()

		if err := subscriber.Run(types.EventTypeRegistrationVote); err != nil {
			log.Errorf("error occured: %v", err)
		}

		sigChan := make(chan os.Signal, 1)

		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan
		log.Infof("signal detected")

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
