package cmd

import (
	"github.com/medibloc/panacea-doracle/types"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	homeDir string
	rootCmd = &cobra.Command{
		Use:   "doracled",
		Short: "doracle daemon",
	}

	nodePrivKeyPath          string
	oraclePrivKeyPath        string
	oraclePubKeyPath         string
	trustedBlockInfoFilePath string
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

	nodePrivKeyPath = filepath.Join(homeDir, types.DefaultNodePrivKeyName)
	oraclePrivKeyPath = filepath.Join(homeDir, types.DefaultOraclePrivKeyName)
	oraclePubKeyPath = filepath.Join(homeDir, types.DefaultOraclePubKeyName)
	trustedBlockInfoFilePath = filepath.Join(homeDir, types.DefaultTrustedBlockInfoName)

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(genOracleKeyCmd)
	rootCmd.AddCommand(verifyReport)
	rootCmd.AddCommand(registerOracleCmd())
	rootCmd.AddCommand(getOracleKeyCmd())
}
