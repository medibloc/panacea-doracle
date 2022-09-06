package service

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	conf        *config.Config
	enclaveInfo *sgx.EnclaveInfo

	oracleAccount *panacea.OracleAccount
	oraclePrivKey []byte

	// queryClient *panacea.QueryClient //TODO: uncomment this
	grpcClient *panacea.GrpcClient
	subscriber *event.PanaceaSubscriber
	txBuilder  *panacea.TxBuilder
}

func New(conf *config.Config, oracleAccount *panacea.OracleAccount) (*Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	oraclePrivKeyPath := filepath.Join(homeDir, ".doracle", types.DefaultOraclePrivKeyName)
	oraclePrivKey, err := sgx.UnsealFromFile(oraclePrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle_priv_key.sealed file: %w", err)
	}

	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to set self-enclave info: %w", err)
	}

	grpcClient, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new gRPC client: %w", err)
	}

	subscriber, err := event.NewSubscriber(conf.Panacea.RPCAddr)
	if err != nil {
		// TODO: close grpcClient
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	txBuilder := panacea.NewTxBuilder(grpcClient)

	return &Service{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		enclaveInfo:   selfEnclaveInfo,
		grpcClient:    grpcClient.(*panacea.GrpcClient),
		subscriber:    subscriber,
		txBuilder:     txBuilder,
	}, nil
}

func (s *Service) StartSubscriptions(events ...event.Event) error {
	return s.subscriber.Run(events...)
}

func (s *Service) Close() error {
	// TODO close query client
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

func (s *Service) OraclePrivKey() []byte {
	return s.oraclePrivKey
}

func (s *Service) EnclaveInfo() *sgx.EnclaveInfo {
	return s.enclaveInfo
}

func (s *Service) GRPCClient() *panacea.GrpcClient {
	return s.grpcClient
}

func (s *Service) TxBuilder() *panacea.TxBuilder {
	return s.txBuilder
}
