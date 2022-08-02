package service

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/store"
)

type Service struct {
	Conf             *config.Config
	Store            store.Storage
	PanaceaClient    panacea.GrpcClientI
	PanaceaWebSocket *event.WebSocket
}

func New(conf *config.Config) (*Service, error) {
	panaceaClient, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Panacea gRPC client: %w", err)
	}

	wsClient, err := event.NewWebSocket(conf)
	if err != nil {
		return nil, err
	}

	return &Service{
		Conf:             conf,
		PanaceaClient:    panaceaClient,
		PanaceaWebSocket: wsClient,
	}, nil
}

func (svc *Service) Close() {
	svc.PanaceaClient.Close()
}
