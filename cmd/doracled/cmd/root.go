package cmd

import (
	"github.com/medibloc/panacea-doracle/types"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	homeDir string
	DbDir   string
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

// init is run automatically when the package is loaded.
func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultAppHomeDir := filepath.Join(userHomeDir, ".doracle")
	DbDir = filepath.Join(defaultAppHomeDir, "data")
	nodePrivKeyPath = filepath.Join(homeDir, types.DefaultNodePrivKeyName)
	oraclePrivKeyPath = filepath.Join(homeDir, types.DefaultOraclePrivKeyName)

	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultAppHomeDir, "application home directory")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(genOracleKeyCmd)
	rootCmd.AddCommand(verifyReport)
	rootCmd.AddCommand(registerOracleCmd())
	rootCmd.AddCommand(getOracleKeyCmd)
}
