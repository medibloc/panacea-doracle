package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto/secp256k1"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"

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
		// get config
		conf, err := config.ReadConfigTOML(getConfigPath())
		if err != nil {
			return fmt.Errorf("failed to read config from file: %w", err)
		}

		// get proper path for oracle key
		oraclePrivKeyPath, oraclePubKeyPath := getKeyPaths(conf.Enclave.Enable)

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
		oraclePrivKey, err := secp256k1.NewPrivKey()
		if err != nil {
			log.Errorf("failed to generate oracle key: %v", err)
			return err
		}

		// seal and store oracle private key
		if err := sgx.SealToFile(oraclePrivKey.Serialize(), oraclePrivKeyPath, conf.Enclave.Enable); err != nil {
			log.Errorf("failed to write %s: %v", oraclePrivKeyPath, err)
			return err
		}

		// generate oracle key remote report
		oraclePubKey := oraclePrivKey.PubKey().SerializeCompressed()
		oracleKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKey[:], conf.Enclave.Enable)
		if err != nil {
			log.Errorf("failed to generate remote report of oracle key: %v", err)
			return err
		}

		// store oracle pub key and its remote report to a file
		err = saveOraclePubKey(oraclePubKey, oracleKeyRemoteReport, oraclePubKeyPath)
		if err != nil {
			log.Errorf("failed to save oracle pub key and its remote report: %v", err)
			return err
		}

		return nil
	},
}

func saveOraclePubKey(oraclePubKey, oracleKeyRemoteReport []byte, path string) error {
	oraclePubKeyData := OraclePubKeyInfo{
		PublicKey:    base64.StdEncoding.EncodeToString(oraclePubKey),
		RemoteReport: base64.StdEncoding.EncodeToString(oracleKeyRemoteReport),
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

func getKeyPaths(enclaveEnabled bool) (string, string) {
	if enclaveEnabled {
		return types.DefaultOraclePrivKeyFilePath, types.DefaultOraclePubKeyFilePath
	} else {
		return filepath.Join(homeDir, types.DefaultOraclePrivKeyName), filepath.Join(homeDir, types.DefaultOraclePubKeyName)
	}
}
