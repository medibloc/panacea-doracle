package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var verifyReport = &cobra.Command{
	Use:   "verify-report [report-file-path]",
	Short: "Verify whether the report was properly generated in the SGX environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// read oracle remote targetReport
		pubKeyInfo, err := readOracleRemoteReport(args[0])
		if err != nil {
			log.Errorf("failed to read remote targetReport: %v", err)
			return err
		}

		if err := verifyPubKeyRemoteReport(*pubKeyInfo); err != nil {
			log.Errorf("failed to verify the public key and its remote report: %v", err)
			return err
		}

		log.Infof("report verification success")
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

func verifyPubKeyRemoteReport(pubKeyInfo OraclePubKeyInfo) error {
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyInfo.PublicKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode oracle public key: %v", err)
	}

	targetReport, err := base64.StdEncoding.DecodeString(pubKeyInfo.RemoteReportBase64)
	if err != nil {
		return fmt.Errorf("failed to decode oracle public key remote report: %v", err)
	}

	// get hash of public key which is used as data
	pubKeyHash := sha256.Sum256(pubKey)

	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return fmt.Errorf("failed to set self-enclave info: %v", err)
	}

	// verify remote report
	if err := sgx.VerifyRemoteReport(targetReport, pubKeyHash[:], *selfEnclaveInfo); err != nil {
		return fmt.Errorf("failed to verify report: %v", err)
	}

	return nil
}
