package panacea

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

type GrpcClientI interface {
	Close() error
	GetInterfaceRegistry() sdk.InterfaceRegistry
	GetChainID() string
	BroadcastTx([]byte) (*tx.BroadcastTxResponse, error)
	GetAccount(string) (authtypes.AccountI, error)
	GetOracleRegistration(string, string) (*oracletypes.OracleRegistration, error)
}

var _ GrpcClientI = (*GrpcClient)(nil)

type GrpcClient struct {
	conn              *grpc.ClientConn
	interfaceRegistry sdk.InterfaceRegistry
	chainID           string
}

// makeInterfaceRegistry
func makeInterfaceRegistry() sdk.InterfaceRegistry {
	interfaceRegistry := sdk.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	return interfaceRegistry
}

func NewGrpcClient(conf *config.Config) (GrpcClientI, error) {
	log.Infof("dialing to Panacea gRPC endpoint: %s", conf.Panacea.GRPCAddr)
	conn, err := grpc.Dial(conf.Panacea.GRPCAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Panacea: %w", err)
	}

	return &GrpcClient{
		conn:              conn,
		interfaceRegistry: makeInterfaceRegistry(),
		chainID:           conf.Panacea.ChainID,
	}, nil
}

func (c *GrpcClient) Close() error {
	log.Info("closing Panacea gRPC connection")
	return c.conn.Close()
}

func (c *GrpcClient) GetInterfaceRegistry() sdk.InterfaceRegistry {
	return c.interfaceRegistry
}

func (c *GrpcClient) GetChainID() string {
	return c.chainID
}

func (c *GrpcClient) GetAccount(panaceaAddr string) (authtypes.AccountI, error) {
	client := authtypes.NewQueryClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: panaceaAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to get account info via grpc: %w", err)
	}

	var acc authtypes.AccountI
	if err := c.interfaceRegistry.UnpackAny(response.GetAccount(), &acc); err != nil {
		return nil, fmt.Errorf("failed to unpack account info: %w", err)
	}
	return acc, nil
}

func (c *GrpcClient) GetOracleRegistration(oracleAddr, uniqueID string) (*oracletypes.OracleRegistration, error) {
	client := oracletypes.NewQueryClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reqMsg := &oracletypes.QueryOracleRegistrationRequest{
		UniqueId: uniqueID,
		Address:  oracleAddr,
	}
	response, err := client.OracleRegistration(ctx, reqMsg)
	if err != nil {
		return nil, err
	}

	return response.OracleRegistration, nil
}

func (c *GrpcClient) BroadcastTx(txBytes []byte) (*tx.BroadcastTxResponse, error) {
	txClient := tx.NewServiceClient(c.conn)

	return txClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		},
	)
}
