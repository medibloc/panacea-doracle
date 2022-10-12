package service

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	conf        *config.Config
	enclaveInfo *sgx.EnclaveInfo

	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey

	queryClient *panacea.QueryClient
	grpcClient  *panacea.GrpcClient
	subscriber  *event.PanaceaSubscriber

	shell *shell.Shell
}

func New(conf *config.Config) (*Service, error) {
	oracleAccount, err := panacea.NewOracleAccount(conf.OracleMnemonic, conf.OracleAccNum, conf.OracleAccIndex)
	if err != nil {
		return nil, err
	}
	oraclePrivKeyBz, err := sgx.UnsealFromFile(conf.AbsOraclePrivKeyPath())
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle_priv_key.sealed file: %w", err)
	}

	oraclePrivKey, _ := crypto.PrivKeyFromBytes(oraclePrivKeyBz)

	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to set self-enclave info: %w", err)
	}

	queryClient, err := panacea.LoadQueryClient(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to load query client: %w", err)
	}

	grpcClient, err := panacea.NewGrpcClient(conf.Panacea.GRPCAddr)
	if err != nil {
		if err := queryClient.Close(); err != nil {
			log.Warn(err)
		}
		return nil, fmt.Errorf("failed to create a new gRPC client: %w", err)
	}

	ipfs := shell.NewShell(conf.Ipfs.IpfsNodeAddr)

	subscriber, err := event.NewSubscriber(conf.Panacea.RPCAddr)
	if err != nil {
		if err := queryClient.Close(); err != nil {
			log.Warn(err)
		}
		if err := grpcClient.Close(); err != nil {
			log.Warn(err)
		}
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	return &Service{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		enclaveInfo:   selfEnclaveInfo,
		queryClient:   queryClient,
		grpcClient:    grpcClient,
		subscriber:    subscriber,
		shell:         ipfs,
	}, nil
}

func (s *Service) StartSubscriptions(events ...event.Event) error {
	return s.subscriber.Run(events...)
}

func (s *Service) Close() error {
	if err := s.queryClient.Close(); err != nil {
		log.Warn(err)
	}
	if err := s.grpcClient.Close(); err != nil {
		log.Warn(err)
	}
	if err := s.subscriber.Close(); err != nil {
		log.Warn(err)
	}

	return nil
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
