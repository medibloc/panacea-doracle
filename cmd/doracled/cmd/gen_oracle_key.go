package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	tos "github.com/tendermint/tendermint/libs/os"
)

// OraclePubKeyInfo is a struct to store oracle public key and its remote report
type OraclePubKeyInfo struct {
	PublicKeyBase64    string `json:"public_key_base64"`
	RemoteReportBase64 string `json:"remote_report_base64"`
}

var genOracleKeyCmd = &cobra.Command{
	Use:   "gen-oracle-key",
	Short: "Generate oracle key and its remote report",
	Long: `Generate a new pair of oracle key and its remote report. 
If the sealed oracle private key exist already, this command will replace the existing one.
So please be cautious in using this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get path for oracle key
		oraclePrivKeyPath := filepath.Join(homeDir, types.DefaultOraclePrivKeyName)
		oraclePubKeyPath := filepath.Join(homeDir, types.DefaultOraclePubKeyName)

		// If there is the existing oracle key, double-check for generating a new oracle key
		if tos.FileExists(oraclePrivKeyPath) {
			buf := bufio.NewReader(os.Stdin)
			ok, err := input.GetConfirmation("This can replace the existing oracle-key.sealed file.\nAre you sure to make a new oracle key?", buf, os.Stderr)

			if err != nil || !ok {
				log.Printf("Oracle key generation is canceled.")
				return err
			}
		}

		// generate a new oracle key
		oraclePrivKey, err := crypto.NewPrivKey()
		if err != nil {
			log.Errorf("failed to generate oracle key: %v", err)
			return err
		}

		// seal and store oracle private key
		if err := sgx.SealToFile(oraclePrivKey.Serialize(), oraclePrivKeyPath); err != nil {
			log.Errorf("failed to write %s: %v", oraclePrivKeyPath, err)
			return err
		}

		// generate oracle key remote report
		oraclePubKey := oraclePrivKey.PubKey().SerializeCompressed()
		oracleKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKey)
		if err != nil {
			log.Errorf("failed to generate remote report of oracle key: %v", err)
			return err
		}

		// store oracle pub key and its remote report to a file
		if err = storeOraclePubKey(oraclePubKey, oracleKeyRemoteReport, oraclePubKeyPath); err != nil {
			log.Errorf("failed to save oracle pub key and its remote report: %v", err)
			return err
		}

		return nil
	},
}

// storeOraclePubKey stores base64-encoded oracle public key and its remote report
func storeOraclePubKey(oraclePubKey, oracleKeyRemoteReport []byte, path string) error {
	oraclePubKeyData := OraclePubKeyInfo{
		PublicKeyBase64:    base64.StdEncoding.EncodeToString(oraclePubKey),
		RemoteReportBase64: base64.StdEncoding.EncodeToString(oracleKeyRemoteReport),
	}

	oraclePubKeyFile, err := json.Marshal(oraclePubKeyData)
	if err != nil {
		return fmt.Errorf("failed to marshal oracle pub key data: %w", err)
	}

	err = ioutil.WriteFile(path, oraclePubKeyFile, 0644)
	if err != nil {
		return fmt.Errorf("failed to write oracle pub key file: %w", err)
	}

	return nil
}
