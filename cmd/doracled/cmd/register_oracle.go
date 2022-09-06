package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/edgelesssys/ego/enclave"
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

func registerOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle",
		Short: "Register an oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			// if node key exists, return error.
			if tos.FileExists(nodePrivKeyPath) {
				return errors.New("node key already exists. If you want to re-generate node key, please delete the node_priv_key.sealed file and retry it")
			}

			// get config
			conf, err := config.ReadConfigTOML(getConfigPath())
			if err != nil {
				return fmt.Errorf("failed to read config.toml: %w", err)
			}

			if err := initLogger(conf); err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}

			// get trusted block information
			trustedBlockInfo, err := getTrustedBlockInfo(cmd)
			if err != nil {
				return fmt.Errorf("failed to get trusted block info: %w", err)
			}

			// get oracle account from mnemonic.
			oracleAccount, err := getOracleAccount(cmd, conf.OracleMnemonic)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// generate node key and its remote report
			nodePubKey, nodePubKeyRemoteReport, err := generateNodeKey()
			if err != nil {
				return fmt.Errorf("failed to generate node key pair: %w", err)
			}

			report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
			uniqueIDStr := base64.StdEncoding.EncodeToString(report.UniqueID)

			// sign and broadcast to Panacea
			msgRegisterOracle := oracletypes.NewMsgRegisterOracle(uniqueIDStr, oracleAccount.GetAddress(), nodePubKey, nodePubKeyRemoteReport, trustedBlockInfo.TrustedBlockHeight, trustedBlockInfo.TrustedBlockHash)

			cli, txBuilder, err := generateGrpcClientAndTxBuilder(conf)
			if err != nil {
				return fmt.Errorf("failed to generate gRPC client and/or Tx builder: %w", err)
			}
			defer func() {
				_ = cli.Close()
			}()

			defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
			txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccount.GetPrivKey(), conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgRegisterOracle)
			if err != nil {
				return fmt.Errorf("failed to generate signed Tx bytes: %w", err)
			}

			resp, err := cli.BroadcastTx(txBytes)
			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %w", err)
			}

			if resp.TxResponse.Code != 0 {
				return fmt.Errorf("register oracle transaction failed: %v", resp.TxResponse.RawLog)
			}

			log.Infof("register-oracle transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

			return nil
		},
	}

	cmd.Flags().Uint32P(flags.FlagAccNum, "a", 0, "Account number of oracle")
	cmd.Flags().Uint32P(flags.FlagIndex, "i", 0, "Address index number for HD derivation of oracle")
	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight)
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHash)

	return cmd
}

// getTrustedBlockInfo gets trusted block height and hash from cmd flags
func getTrustedBlockInfo(cmd *cobra.Command) (*panacea.TrustedBlockInfo, error) {
	trustedBlockHeight, err := cmd.Flags().GetInt64(flags.FlagTrustedBlockHeight)
	if err != nil {
		return nil, err
	}
	if trustedBlockHeight == 0 {
		return nil, fmt.Errorf("trusted block height cannot be zero")
	}

	trustedBlockHashStr, err := cmd.Flags().GetString(flags.FlagTrustedBlockHash)
	if err != nil {
		return nil, err
	}
	if trustedBlockHashStr == "" {
		return nil, fmt.Errorf("trusted block hash cannot be empty")
	}

	trustedBlockHash, err := base64.StdEncoding.DecodeString(trustedBlockHashStr)
	if err != nil {
		return nil, err
	}

	return &panacea.TrustedBlockInfo{
		TrustedBlockHeight: trustedBlockHeight,
		TrustedBlockHash:   trustedBlockHash,
	}, nil
}

// getOracleAccount gets an oracle account from mnemonic.
// The account is equal to one that is registered as validator.
// You can set account number and index optionally.
// The default value is 0 for both account number and index
func getOracleAccount(cmd *cobra.Command, mnemonic string) (*panacea.OracleAccount, error) {
	accNum, err := cmd.Flags().GetUint32(flags.FlagAccNum)
	if err != nil {
		return &panacea.OracleAccount{}, err
	}

	index, err := cmd.Flags().GetUint32(flags.FlagIndex)
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
