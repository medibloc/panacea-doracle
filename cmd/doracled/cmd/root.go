package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	NodePrivKeyFilePath   = "/data/node_priv_key.sealed"
	OraclePrivKeyFilePath = "/data/.doracle/oracle_priv_key.sealed"
	OraclePubKeyFilePath  = "/data/.doracle/oracle_pub_key.json"
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
