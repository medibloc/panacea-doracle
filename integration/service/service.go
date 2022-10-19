package service

import (
	"context"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/store"
	dbm "github.com/tendermint/tm-db"
)

type TestServiceWithoutSGX struct {
	conf          *config.Config
	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient *panacea.QueryClient
	grpcClient  *panacea.GrpcClient
	subscriber  *event.PanaceaSubscriber
	ipfs        *store.Ipfs
}

func NewTestServiceWithoutSGX(conf *config.Config, info *panacea.TrustedBlockInfo) (*TestServiceWithoutSGX, error) {
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

	ipfs := store.NewIpfs(conf.Ipfs.IpfsNodeAddr)

	return &TestServiceWithoutSGX{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		queryClient:   queryClient,
		grpcClient:    grpcClient,
		subscriber:    panaceaSubscriber,
		ipfs:          ipfs,
	}, nil
}

func (s *TestServiceWithoutSGX) Config() *config.Config {
	return s.conf
}

func (s *TestServiceWithoutSGX) OracleAcc() *panacea.OracleAccount {
	return s.oracleAccount
}

func (s *TestServiceWithoutSGX) OraclePrivKey() *btcec.PrivateKey {
	return s.oraclePrivKey
}

func (s *TestServiceWithoutSGX) EnclaveInfo() *sgx.EnclaveInfo {
	return nil
}

func (s *TestServiceWithoutSGX) GRPCClient() *panacea.GrpcClient {
	return s.grpcClient
}

func (s *TestServiceWithoutSGX) QueryClient() *panacea.QueryClient {
	return s.queryClient
}
