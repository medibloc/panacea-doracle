package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var verifyReport = &cobra.Command{
	Use:   "verify-report [report-file-path] [unique-id]",
	Short: "Verify whether the report was properly generated in the SGX environment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// read oracle remote report
		pubKeyInfo, err := readOracleRemoteReport(args[0])
		if err != nil {
			log.Errorf("failed to read remote report: %v", err)
			return err
		}

		if len(pubKeyInfo.RemoteReport) == 0 {
			log.Errorf("invalid report: report is empty")
			return nil
		}

		pubKey, err := base64.StdEncoding.DecodeString(pubKeyInfo.PublicKey)
		if err != nil {
			log.Errorf("failed to decode oracle public key: %v", err)
			return err
		}

		report, err := base64.StdEncoding.DecodeString(pubKeyInfo.RemoteReport)
		if err != nil {
			log.Errorf("failed to decode oracle remote report: %v", err)
			return err
		}

		// get hash of public key which is used as data
		pubKeyHash := sha256.Sum256(pubKey)
		if err := sgx.VerifyRemoteReport(report, pubKeyHash[:], args[1]); err != nil {
			log.Errorf("failed to verify report: %v", err)
			return err
		}

		log.Printf("report verification success")
		return nil
	},
}

func readOracleRemoteReport(filename string) (*OraclePubKeyInfo, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var pubKeyInfo OraclePubKeyInfo

	if err := json.Unmarshal(file, &pubKeyInfo); err != nil {
		return nil, err
	}

	return &pubKeyInfo, nil
}
