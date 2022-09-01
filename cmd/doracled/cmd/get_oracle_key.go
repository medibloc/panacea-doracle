package cmd

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
)

var getOracleKeyCmd = &cobra.Command{
	Use:   "get-oracle-key",
	Short: "Get a shared oracle private key",
	Long: `Get a shared oracle private key from Panacea.
The encrypted oracle private key can only be decrypted using the node key, which is generated in SGX-enabled environment.
The node key must be the same with the one stored in OracleRegistration KV store in Panacea.
After decrypted, the oracle private key is sealed and stored as a file named oracle_priv_key.sealed securely.
This oracle private key can also be accessed in SGX-enabled environment using the promised binary.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// if there is a node key already, return error
		if !tos.FileExists(nodePrivKeyPath) {
			return errors.New("no node_priv_key.sealed file")
		}

		// get config
		conf, err := config.ReadConfigTOML(getConfigPath())
		if err != nil {
			return fmt.Errorf("failed to read config.toml: %w", err)
		}

		// get existing node key
		nodePrivKeyBz, err := sgx.UnsealFromFile(nodePrivKeyPath)
		if err != nil {
			return fmt.Errorf("failed to unseal node_priv_key.sealed file: %w", err)
		}
		nodePrivKey, nodePubKey := crypto.PrivKeyFromBytes(nodePrivKeyBz)

		// get unique ID
		selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
		if err != nil {
			return fmt.Errorf("failed to get self enclave info: %w", err)
		}
		uniqueID := hex.EncodeToString(selfEnclaveInfo.UniqueID)

		// get oracle account from mnemonic.
		oracleAccount, err := getOracleAccount(cmd, conf.OracleMnemonic)
		if err != nil {
			return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
		}

		// TODO: replace to use query client
		// get OracleRegistration from Panacea
		cli, err := panacea.NewGrpcClient(conf)
		if err != nil {
			return fmt.Errorf("failed to create gRPC client: %w", err)
		}
		defer func() {
			_ = cli.Close()
		}()

		oracleRegistration, err := cli.GetOracleRegistration(oracleAccount.GetAddress(), uniqueID)
		if err != nil {
			return fmt.Errorf("failed to get oracle registration from Panacea: %w", err)
		}

		// check if the same node key is used for oracle registration
		if !bytes.Equal(oracleRegistration.NodePubKey, nodePubKey.SerializeCompressed()) {
			return errors.New("the existing node key is different from the one used in oracle registration. if you want to re-request RegisterOracle, delete the existing node_priv_key.sealed file and rerun register-oracle cmd")
		}

		return getOraclePrivKey(oracleRegistration, nodePrivKey)
	},
}

// getOraclePrivKey handles OracleRegistration differently depending on the status of oracle registration
func getOraclePrivKey(oracleRegistration *oracletypes.OracleRegistration, nodePrivKey *btcec.PrivateKey) error {
	switch oracleRegistration.Status {
	case oracletypes.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD:
		return errors.New("voting is currently in progress")

	case oracletypes.ORACLE_REGISTRATION_STATUS_PASSED:
		// if exists, no need to do get-oracle-key cmd.
		if tos.FileExists(oraclePrivKeyPath) {
			return errors.New("the oracle private key already exists")
		}

		// else, get encryptedOraclePrivKey from Panacea and decrypt and SealToFile it
		oraclePrivKey, err := crypto.Decrypt(nodePrivKey, oracleRegistration.EncryptedOraclePrivKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt the EncryptedOraclePrivKey: %w", err)
		}

		if err := sgx.SealToFile(oraclePrivKey, oraclePrivKeyPath); err != nil {
			return fmt.Errorf("failed to seal to file: %w", err)
		}

		log.Info("oracle private key is retrieved successfully")
		return nil

	case oracletypes.ORACLE_REGISTRATION_STATUS_REJECTED:
		return errors.New("the request for oracle registration is rejected. please delete the existing node key and retry with a new one")

	default:
		return errors.New("invalid oracle registration status")
	}
}
