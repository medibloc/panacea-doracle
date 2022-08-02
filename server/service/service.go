package service

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/store"
)

type Service struct {
	Conf             *config.Config
	Store            store.Storage
	PanaceaClient    panacea.GrpcClientI
}

func New(conf *config.Config) (*Service, error) {
	panaceaClient, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Panacea gRPC client: %w", err)
	}

	return &Service{
		Conf:             conf,
		PanaceaClient:    panaceaClient,
	}, nil
}

func (svc *Service) Close() {
	svc.PanaceaClient.Close()
}
