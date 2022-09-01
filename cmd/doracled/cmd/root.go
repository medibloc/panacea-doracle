package cmd

import (
	"context"
	"fmt"
	"github.com/medibloc/panacea-doracle/client"
	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"

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

	ctx := client.Context{}
	ctx = ctx.WithHomeDir(defaultAppHomeDir)
	if err := rootCmd.ExecuteContext(context.WithValue(context.Background(), client.ContextKey, &ctx)); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(genOracleKeyCmd)
	rootCmd.AddCommand(verifyReport)
	rootCmd.AddCommand(registerOracleCmd())
	rootCmd.AddCommand(getOracleKeyCmd)
}
