package service

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
)

type Service struct {
	Conf          *config.Config
	OracleAccount *panacea.OracleAccount
	OraclePrivKey *btcec.PrivateKey
	UniqueID      string
	HomeDir       string

	QueryClient *panacea.QueryClient
	GrpcClient  *panacea.GrpcClient
}

func New(conf *config.Config, homeDir string) (*Service, error) {
	oraclePrivKeyPath := types.GetOraclePrivKeyPath(homeDir)
	oraclePrivKeyBz, err := sgx.UnsealFromFile(oraclePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle_priv_key.sealed file: %w", err)
	}
	oraclePrivKey, _ := crypto.PrivKeyFromBytes(oraclePrivKeyBz)

	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to set self-enclave info: %w", err)
	}

	grpcClient, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gRPC client: %w", err)
	}

	return &Service{
		Conf:          conf,
		OraclePrivKey: oraclePrivKey,
		UniqueID:      hex.EncodeToString(selfEnclaveInfo.UniqueID),
		GrpcClient:    grpcClient.(*panacea.GrpcClient),
		HomeDir:       homeDir,
	}, nil
}

func (s Service) Close() error {
	// TODO close query client
	if err := s.GrpcClient.Close(); err != nil {
		return err
	}

	return nil
}
