package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/client/flags"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/libs/os"
	"io"

	"github.com/edgelesssys/ego/enclave"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/spf13/cobra"
)

func upgradeOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade-oracle",
		Short: "Upgrade an oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadConfigFromHome(cmd)
			if err != nil {
				return err
			}
			ctx := context.Background()

			// if node key exists, return error.
			nodePrivKeyPath := conf.AbsNodePrivKeyPath()
			if os.FileExists(nodePrivKeyPath) {
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

			msg := oracletypes.NewMsgUpgradeOracle(
				uniqueID,
				oracleAccount.GetAddress(),
				nodePubKey,
				nodePubKeyRemoteReport,
				trustedBlockInfo.TrustedBlockHeight,
				trustedBlockInfo.TrustedBlockHash,
				nonce,
			)

			txBuilder := panacea.NewTxBuilder(*queryClient)
			cli, err := panacea.NewGrpcClient(conf.Panacea.GRPCAddr)
			if err != nil {
				return fmt.Errorf("failed to generate gRPC client: %w", err)
			}
			defer cli.Close()

			defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
			txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccount.GetPrivKey(), conf.Panacea.DefaultGasLimit, defaultFeeAmount, msg)
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

			log.Infof("upgrade-oracle transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

			return nil
		},
	}

	cmd.Flags().Int64(flags.FlagTrustedBlockHeight, 0, "Trusted block height")
	cmd.Flags().String(flags.FlagTrustedBlockHash, "", "Trusted block hash")
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHeight)
	_ = cmd.MarkFlagRequired(flags.FlagTrustedBlockHash)

	return cmd
}
