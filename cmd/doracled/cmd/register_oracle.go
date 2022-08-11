package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path/filepath"
)

const (
	FlagTrustedBlockHeight = "trusted-block-height"
	FlagTrustedBlockHash   = "trusted-block-hash"
	FlagAccNum             = "acc-num"
	FlagIndex              = "index"
)

func RegisterOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle",
		Short: "Register an oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get config
			conf, err := config.ReadConfigTOML(getConfigPath())
			if err != nil {
				return err
			}

			if err := initLogger(conf); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}

			// get trusted block information
			height, hash, err := getTrustedBlockInfo(cmd)
			if err != nil {
				log.Errorf("failed to get trusted block info: %v", err)
				return err
			}

			// get oracle account from mnemonic.
			oracleAccount, err := getOracleAccount(cmd, conf.OracleMnemonic)
			if err != nil {
				log.Errorf("failed to get oracle account from mnemonic: %v", err)
				return err
			}

			// generate node key and its remote report
			nodePubKey, nodePubKeyRemoteReport, err := generateNodeKey()
			if err != nil {
				log.Errorf("failed to generate node key pair: %v", err)
				return err
			}

			report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
			uniqueIDStr := base64.StdEncoding.EncodeToString(report.UniqueID)

			// sign and broadcast to Panacea
			msgRegisterOracle := oracletypes.NewMsgRegisterOracle(uniqueIDStr, oracleAccount.GetAddress(), nodePubKey, nodePubKeyRemoteReport, height, hash)

			cli, txBuilder, err := generateGrpcClientAndTxBuilder(conf)
			if err != nil {
				log.Errorf("failed to generate gRPC client and/or Tx builder: %v", err)
				return err
			}

			defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)

			txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccount.GetPrivKey(), conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgRegisterOracle)
			if err != nil {
				log.Errorf("failed to generate signed Tx bytes: %v", err)
				return err
			}

			if _, err := cli.BroadcastTx(txBytes); err != nil {
				log.Errorf("failed to broadcast transaction: %v", err)
				return err
			}

			return nil
		},
	}

	cmd.Flags().Uint32P(FlagAccNum, "a", 0, "Account number of oracle")
	cmd.Flags().Uint32P(FlagIndex, "i", 0, "Address index number for HD derivation of oracle")
	cmd.Flags().Int64(FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(FlagTrustedBlockHash, "", "Trusted block hash")
	_ = cmd.MarkFlagRequired(FlagTrustedBlockHeight)
	_ = cmd.MarkFlagRequired(FlagTrustedBlockHash)

	return cmd
}

// getTrustedBlockInfo gets trusted block height and hash from cmd flags
func getTrustedBlockInfo(cmd *cobra.Command) (int64, []byte, error) {
	trustedBlockHeight, err := cmd.Flags().GetInt64(FlagTrustedBlockHeight)
	if err != nil {
		return 0, nil, err
	}
	if trustedBlockHeight == 0 {
		return 0, nil, fmt.Errorf("trusted block height cannot be zero")
	}

	trustedBlockHashStr, err := cmd.Flags().GetString(FlagTrustedBlockHash)
	if err != nil {
		return 0, nil, err
	}
	if trustedBlockHashStr == "" {
		return 0, nil, fmt.Errorf("trusted block hash cannot be empty")
	}

	trustedBlockHash, err := base64.StdEncoding.DecodeString(trustedBlockHashStr)
	if err != nil {
		return 0, nil, err
	}

	return trustedBlockHeight, trustedBlockHash, nil
}

// getOracleAccount gets an oracle account from mnemonic.
// The account is equal to one that is registered as validator.
// You can set account number and index optionally.
// The default value is 0 for both account number and index
func getOracleAccount(cmd *cobra.Command, mnemonic string) (*panacea.OracleAccount, error) {
	accNum, err := cmd.Flags().GetUint32(FlagAccNum)
	if err != nil {
		return &panacea.OracleAccount{}, err
	}

	index, err := cmd.Flags().GetUint32(FlagIndex)
	if err != nil {
		return &panacea.OracleAccount{}, err
	}

	oracleAccount, err := panacea.NewOracleAccount(mnemonic, accNum, index)
	if err != nil {
		return &panacea.OracleAccount{}, err
	}

	return oracleAccount, nil
}

// generateNodeKey generates random node key and its remote report
// And the generated private key is sealed and stored
func generateNodeKey() ([]byte, []byte, error) {
	nodePrivKey, err := crypto.NewPrivKey()
	if err != nil {
		return nil, nil, err
	}

	nodePrivKeyPath := filepath.Join(homeDir, types.DefaultNodePrivKeyName)
	if err := sgx.SealToFile(nodePrivKey.Serialize(), nodePrivKeyPath); err != nil {
		return nil, nil, err
	}

	nodePubKey := nodePrivKey.PubKey().SerializeCompressed()
	oraclePubKeyHash := sha256.Sum256(nodePubKey)
	nodeKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKeyHash[:])
	if err != nil {
		return nil, nil, err
	}

	return nodePubKey, nodeKeyRemoteReport, nil
}

// generateGrpcClientAndTxBuilder generates gRPC client and TxBuilder
func generateGrpcClientAndTxBuilder(conf *config.Config) (panacea.GrpcClientI, *panacea.TxBuilder, error) {
	cli, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, nil, err
	}

	return cli, panacea.NewTxBuilder(cli), nil
}
