package cmd

import (
	"github.com/medibloc/panacea-doracle/crypto/secp256k1"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var genOracleKeyCmd = &cobra.Command{
	Use:   "gen-oracle-key",
	Short: "Generate oracle key and its remote report",
	Long: `Generate a new pair of oracle key and its remote report. 
If the sealed oracle private key exist already, this cli will replace the existing one.
So please be cautious in using this cli.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// generate random oracle key
		oraclePrivKey, err := secp256k1.NewPrivKey()
		if err != nil {
			log.Fatalf("failed to generate oracle key: %v", err)
		}

		// seal and store oracle private key
		if err := sgx.SealToFile(oraclePrivKey.Serialize(), OracleKeyFilePath); err != nil {
			log.Fatalf("failed to save oracle key: %v", err)
		}

		// generate oracle key remote report
		oraclePubKey := oraclePrivKey.PubKey().SerializeCompressed()
		report, err := sgx.GenerateRemoteReport(oraclePubKey[:])
		if err != nil {
			return err
		}

		log.Info(report)
		// print? oracle pub key and remote report

		return nil
	},
}
