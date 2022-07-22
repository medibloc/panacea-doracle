package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	NodeKeyFilePath   = "/data/node-key.sealed"
	OracleKeyFilePath = "/data/oracle-key.sealed"
)

var (
	homeDir string
	rootCmd = &cobra.Command{
		Use:   "doracled",
		Short: "doracle daemon",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

// init is run automatically when the package is loaded.
func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultAppHomeDir := filepath.Join(userHomeDir, ".doracle")

	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultAppHomeDir, "application home directory")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(genOracleKeyCmd)
}
