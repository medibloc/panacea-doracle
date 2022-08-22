package service

import (
	"encoding/base64"
	"fmt"
	"github.com/medibloc/panacea-doracle/client/flags"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/spf13/cobra"
)

type Service struct {
	Conf          *config.Config
	OracleAccount *panacea.OracleAccount
	TrustedHeight int64
	TrustedHash   []byte
	PanaceaClient panacea.GrpcClientI
}

func New(cmd *cobra.Command, conf *config.Config) (*Service, error) {
	// get trusted block information
	height, hash, err := getTrustedBlockInfo(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted block info: %w", err)
	}

	// get oracle account from mnemonic.
	oracleAccount, err := getOracleAccount(cmd, conf.OracleMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle account from mnemonic: %w", err)
	}

	panaceaClient, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Panacea gRPC client: %w", err)
	}

	return &Service{
		Conf:          conf,
		OracleAccount: oracleAccount,
		TrustedHeight: height,
		TrustedHash:   hash,
		PanaceaClient: panaceaClient,
	}, nil
}

func (svc *Service) Close() {
	_ = svc.PanaceaClient.Close()
}

// getTrustedBlockInfo gets trusted block height and hash from cmd flags
func getTrustedBlockInfo(cmd *cobra.Command) (int64, []byte, error) {
	trustedBlockHeight, err := cmd.Flags().GetInt64(flags.FlagTrustedBlockHeight)
	if err != nil {
		return 0, nil, err
	}
	if trustedBlockHeight == 0 {
		return 0, nil, fmt.Errorf("trusted block height cannot be zero")
	}

	trustedBlockHashStr, err := cmd.Flags().GetString(flags.FlagTrustedBlockHash)
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
