package service

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
)

type Service struct {
	Conf             *config.Config
	OracleAccount    *panacea.OracleAccount
	TrustedBlockInfo *panacea.TrustedBlockInfo
	PanaceaClient    panacea.GrpcClientI
}

func New(conf *config.Config, oracleAccount *panacea.OracleAccount, trustedBlockInfo *panacea.TrustedBlockInfo) (*Service, error) {
	panaceaClient, err := panacea.NewGrpcClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Panacea gRPC client: %w", err)
	}

	return &Service{
		Conf:             conf,
		OracleAccount:    oracleAccount,
		TrustedBlockInfo: trustedBlockInfo,
		PanaceaClient:    panaceaClient,
	}, nil
}

func (svc *Service) Close() {
	_ = svc.PanaceaClient.Close()
}
