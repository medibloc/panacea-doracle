package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/client/flags"
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
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()

			// if node key exists, return error.
			nodePrivKeyPath := conf.AbsNodePrivKeyPath()
			if tos.FileExists(nodePrivKeyPath) {
				return errors.New("node key already exists. If you want to re-generate node key, please delete the node_priv_key.sealed file and retry it")
			}

			// get trusted block information
			trustedBlockInfo, err := getTrustedBlockInfo(cmd)
			if err != nil {
				return fmt.Errorf("failed to get trusted block info: %w", err)
			}

			// initialize query client using trustedBlockInfo
			queryClient, err := panacea.NewQueryClient(ctx, conf, *trustedBlockInfo)
			if err != nil {
				return fmt.Errorf("failed to initialize QueryClient: %w", err)
			}
			defer queryClient.Close()

			// get oracle account from mnemonic.
			oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
			if err != nil {
				return fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
			}

			// generate node key and its remote report
			nodePubKey, nodePubKeyRemoteReport, err := generateNodeKey(nodePrivKeyPath)
			if err != nil {
				return fmt.Errorf("failed to generate node key pair: %w", err)
			}

			report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
			uniqueID := hex.EncodeToString(report.UniqueID)

			nonce := make([]byte, 12)
			_, err = io.ReadFull(rand.Reader, nonce)
			if err != nil {
				return fmt.Errorf("failed to make nonce: %w", err)
			}

			// sign and broadcast to Panacea
			msgRegisterOracle := oracletypes.NewMsgRegisterOracle(uniqueID, oracleAccount.GetAddress(), nodePubKey, nodePubKeyRemoteReport, trustedBlockInfo.TrustedBlockHeight, trustedBlockInfo.TrustedBlockHash, nonce)

			txBuilder := panacea.NewTxBuilder(*queryClient)
			cli, err := panacea.NewGrpcClient(conf.Panacea.GRPCAddr)
			if err != nil {
				return fmt.Errorf("failed to generate gRPC client: %w", err)
			}
			defer cli.Close()

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

	trustedBlockHash, err := hex.DecodeString(trustedBlockHashStr)
	if err != nil {
		return nil, err
	}

	return &panacea.TrustedBlockInfo{
		TrustedBlockHeight: trustedBlockHeight,
		TrustedBlockHash:   trustedBlockHash,
	}, nil
}

// generateNodeKey generates random node key and its remote report
// And the generated private key is sealed and stored
func generateNodeKey(nodePrivKeyPath string) ([]byte, []byte, error) {
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
