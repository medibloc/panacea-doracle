package panacea

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GrpcClientI interface {
	Close() error
}

var _ GrpcClientI = (*GrpcClient)(nil)

type GrpcClient struct {
	conn              *grpc.ClientConn
	interfaceRegistry sdk.InterfaceRegistry
}

// makeInterfaceRegistry
func makeInterfaceRegistry() sdk.InterfaceRegistry {
	interfaceRegistry := sdk.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	aoltypes.RegisterInterfaces(interfaceRegistry)
	return interfaceRegistry
}

func NewGrpcClient(conf *config.Config) (GrpcClientI, error) {
	log.Infof("dialing to Panacea gRPC endpoint: %s", conf.Panacea.GRPCAddr)
	conn, err := grpc.Dial(conf.Panacea.GRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Panacea: %w", err)
	}

	return &GrpcClient{
		conn:              conn,
		interfaceRegistry: makeInterfaceRegistry(),
	}, nil
}

func (c *GrpcClient) Close() error {
	log.Info("closing Panacea gRPC connection")
	return c.conn.Close()
}
