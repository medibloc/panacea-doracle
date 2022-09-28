package service

import (
	"context"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	dbm "github.com/tendermint/tm-db"
)

type TestService struct {
	conf          *config.Config
	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient *panacea.QueryClient
	grpcClient  *panacea.GrpcClient
	subscriber  *event.PanaceaSubscriber
}

func NewTestService(conf *config.Config, info *panacea.TrustedBlockInfo) (*TestService, error) {
	oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
	if err != nil {
		return nil, err
	}

	oraclePrivKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}

	queryClient, err := panacea.NewQueryClientWithDB(context.Background(), conf, info, dbm.NewMemDB())
	if err != nil {
		return nil, err
	}

	grpcClient, err := panacea.NewGrpcClient(conf.Panacea.GRPCAddr)
	if err != nil {
		return nil, err
	}

	panaceaSubscriber, err := event.NewSubscriber(conf.Panacea.RPCAddr)
	if err != nil {
		return nil, err
	}

	return &TestService{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		queryClient:   queryClient,
		grpcClient:    grpcClient,
		subscriber:    panaceaSubscriber,
	}, nil
}

func (s *TestService) Config() *config.Config {
	return s.conf
}

func (s *TestService) OracleAcc() *panacea.OracleAccount {
	return s.oracleAccount
}

func (s *TestService) OraclePrivKey() *btcec.PrivateKey {
	return s.oraclePrivKey
}

func (s *TestService) EnclaveInfo() *sgx.EnclaveInfo {
	return nil
}

func (s *TestService) GRPCClient() *panacea.GrpcClient {
	return s.grpcClient
}

func (s *TestService) QueryClient() *panacea.QueryClient {
	return s.queryClient
}
