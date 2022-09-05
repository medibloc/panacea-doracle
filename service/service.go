package service

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	conf          *config.Config
	oracleAccount *panacea.OracleAccount
	oraclePrivKey *btcec.PrivateKey
	uniqueID      string

	// queryClient *panacea.QueryClient //TODO: uncomment this
	grpcClient *panacea.GrpcClient
	subscriber *event.PanaceaSubscriber
}

func New(conf *config.Config, oracleAccount *panacea.OracleAccount) (*Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	oraclePrivKeyPath := filepath.Join(homeDir, ".doracle", types.DefaultOraclePrivKeyName)
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

	subscriber, err := event.NewSubscriber(conf.Panacea.WSAddr)
	if err != nil {
		// TODO: close grpcClient
		return nil, fmt.Errorf("failed to init subscriber: %w", err)
	}

	svc := &Service{
		conf:          conf,
		oracleAccount: oracleAccount,
		oraclePrivKey: oraclePrivKey,
		uniqueID:      hex.EncodeToString(selfEnclaveInfo.UniqueID),
		grpcClient:    grpcClient.(*panacea.GrpcClient),
		subscriber:    subscriber,
	}

	if err := svc.startSubscriptions(); err != nil {
		return nil, fmt.Errorf("failed to start subscriptions: %w", err)
	}

	return svc, nil
}

func (s *Service) startSubscriptions() error {
	return s.subscriber.Run(
		event.NewRegisterOracleEvent(s),
	)
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

func (s *Service) GRPCClient() *panacea.GrpcClient {
	return s.grpcClient
}

func (s *Service) UniqueID() string {
	return s.uniqueID
}
