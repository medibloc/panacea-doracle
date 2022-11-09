package service

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/ipfs"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	dbm "github.com/tendermint/tm-db"
)

type Service struct {
	conf        *config.Config
	enclaveInfo *sgx.EnclaveInfo

	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient *panacea.QueryClient
	grpcClient  *panacea.GrpcClient
	subscriber  *event.PanaceaSubscriber
	ipfs        *ipfs.Ipfs
}

func NewTestService(conf *config.Config, info *panacea.TrustedBlockInfo, enclaveInfo *sgx.EnclaveInfo) (*Service, error) {
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

	ipfs := ipfs.NewIpfs(conf.Ipfs.IpfsNodeAddr)

	return &Service{
		conf:          conf,
		enclaveInfo:   enclaveInfo,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		queryClient:   queryClient,
		grpcClient:    grpcClient,
		subscriber:    panaceaSubscriber,
		ipfs:          ipfs,
	}, nil
}

func NewTestServiceWithoutSGX(conf *config.Config, info *panacea.TrustedBlockInfo) (*Service, error) {
	return NewTestService(conf, info, nil)
}

func (s *Service) Config() *config.Config {
	return s.conf
}

func (s *Service) OracleAcc() *panacea.OracleAccount {
	return s.oracleAccount
}

func (s *Service) OraclePrivKey() *btcec.PrivateKey {
	return s.oraclePrivKey
}

func (s *Service) EnclaveInfo() *sgx.EnclaveInfo {
	return s.enclaveInfo
}

func (s *Service) GRPCClient() *panacea.GrpcClient {
	return s.grpcClient
}

func (s *Service) QueryClient() *panacea.QueryClient {
	return s.queryClient
}

func (s *Service) Ipfs() *ipfs.Ipfs {
	return s.ipfs
}

func (s *Service) BroadcastTx(txBytes []byte) (int64, string, error) {
	resp, err := s.GRPCClient().BroadcastTx(txBytes)
	if err != nil {
		return 0, "", fmt.Errorf("broadcast transaction failed. txBytes(%v)", txBytes)
	}

	if resp.TxResponse.Code != 0 {
		return 0, "", fmt.Errorf("transaction failed: %v", resp.TxResponse.RawLog)
	}

	return resp.TxResponse.Height, resp.TxResponse.TxHash, nil
}
