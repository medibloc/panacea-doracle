package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tos "github.com/tendermint/tendermint/libs/os"
)

func getOracleKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-oracle-key",
		Short: "Get a shared oracle private key",
		Long: `Get a shared oracle private key from Panacea.
		The encrypted oracle private key can only be decrypted using the node key, which is generated in SGX-enabled environment.
		The node key must be the same with the one stored in OracleRegistration KV store in Panacea.
		After decrypted, the oracle private key is sealed and stored as a file named oracle_priv_key.sealed securely.
		This oracle private key can also be accessed in SGX-enabled environment using the promised binary.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}

			ctx := context.Background()

			// if the node key does not exist, return error
			nodePrivKeyPath := conf.AbsNodePrivKeyPath()
			if !tos.FileExists(nodePrivKeyPath) {
				return errors.New("no node_priv_key.sealed file")
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
			uniqueID := selfEnclaveInfo.UniqueIDHex()

			// get oracle account from mnemonic.
			oracleAccount, err := getOracleAccount(cmd, conf.OracleMnemonic)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// get OracleRegistration from Panacea
			queryClient, err := panacea.LoadQueryClient(ctx, conf)
			if err != nil {
				return fmt.Errorf("failed to get queryClient: %w", err)
			}
			defer queryClient.Close()

			oracleRegistration, err := queryClient.GetOracleRegistration(oracleAccount.GetAddress(), uniqueID)
			if err != nil {
				return fmt.Errorf("failed to get oracle registration from Panacea: %w", err)
			}

			// check if the same node key is used for oracle registration
			if !bytes.Equal(oracleRegistration.NodePubKey, nodePubKey.SerializeCompressed()) {
				return errors.New("the existing node key is different from the one used in oracle registration. if you want to re-request RegisterOracle, delete the existing node_priv_key.sealed file and rerun register-oracle cmd")
			}

			oraclePublicKey, err := queryClient.GetOracleParamsPublicKey()
			if err != nil {
				return err
			}

			return getOraclePrivKey(conf, oracleRegistration, nodePrivKey, oraclePublicKey)
		},
	}
	cmd.Flags().Uint32P(flags.FlagAccNum, "a", 0, "Account number of oracle")
	cmd.Flags().Uint32P(flags.FlagIndex, "i", 0, "Address index number for HD derivation of oracle")

	return cmd
}

// getOraclePrivKey handles OracleRegistration differently depending on the status of oracle registration
func getOraclePrivKey(conf *config.Config, oracleRegistration *oracletypes.OracleRegistration, nodePrivKey *btcec.PrivateKey, oraclePubKey *btcec.PublicKey) error {
	switch oracleRegistration.Status {
	case oracletypes.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD:
		return errors.New("voting is currently in progress")

	case oracletypes.ORACLE_REGISTRATION_STATUS_PASSED:
		// if exists, no need to do get-oracle-key cmd.
		oraclePrivKeyPath := conf.AbsOraclePrivKeyPath()
		if tos.FileExists(oraclePrivKeyPath) {
			return errors.New("the oracle private key already exists")
		}

		shareKey := crypto.ShareKey(nodePrivKey, oraclePubKey)

		oraclePrivKey, err := crypto.DecryptWithAES256(shareKey, oracleRegistration.Nonce, oracleRegistration.EncryptedOraclePrivKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt the encrypted oracle private key: %w", err)
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
