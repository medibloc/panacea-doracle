package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/medibloc/panacea-doracle/crypto/secp256k1"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	tos "github.com/tendermint/tendermint/libs/os"
)

type OraclePubKeyInfo struct {
	PublicKey    string `json:"public_key"`
	RemoteReport string `json:"remote_report"`
}

var genOracleKeyCmd = &cobra.Command{
	Use:   "gen-oracle-key",
	Short: "Generate oracle key and its remote report",
	Long: `Generate a new pair of oracle key and its remote report. 
If the sealed oracle private key exist already, this command will replace the existing one.
So please be cautious in using this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If there is the existing oracle key, double-check for generating a new oracle key
		if tos.FileExists(types.OraclePrivKeyFilePath) {
			buf := bufio.NewReader(os.Stdin)
			ok, err := input.GetConfirmation("This can replace the existing oracle-key.sealed file.\nAre you sure to make a new oracle key?", buf, os.Stderr)

			if err != nil || !ok {
				log.Printf("Oracle key generation is canceled.")
				return err
			}
		}

		// randomly generate a new oracle key
		oraclePrivKey, err := secp256k1.NewPrivKey()
		if err != nil {
			log.Errorf("failed to generate oracle key: %v", err)
			return err
		}

		// seal and store oracle private key
		if err := sgx.SealToFile(oraclePrivKey.Serialize(), types.OraclePrivKeyFilePath); err != nil {
			log.Errorf("failed to save oracle key: %v", err)
			return err
		}

		// generate oracle key remote report
		oraclePubKey := oraclePrivKey.PubKey().SerializeCompressed()
		oracleKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKey[:])
		if err != nil {
			log.Errorf("failed to generate remote report of oracle key: %v", err)
			return err
		}

		// store oracle pub key and its remote report to a file
		err = saveOraclePubKey(oraclePubKey, oracleKeyRemoteReport)
		if err != nil {
			log.Errorf("failed to save oracle pub key and its remote report: %v", err)
			return err
		}

		return nil
	},
}

func saveOraclePubKey(oraclePubKey, oracleKeyRemoteReport []byte) error {
	oraclePubKeyData := OraclePubKeyInfo{
		PublicKey:    base64.StdEncoding.EncodeToString(oraclePubKey),
		RemoteReport: base64.StdEncoding.EncodeToString(oracleKeyRemoteReport),
	}

	oraclePubKeyFile, err := json.Marshal(oraclePubKeyData)
	if err != nil {
		return fmt.Errorf("failed to marshal oracle pub key data: %w", err)
	}

	err = ioutil.WriteFile(types.OraclePubKeyFilePath, oraclePubKeyFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write oracle pub key file: %w", err)
	}

	return nil
}
