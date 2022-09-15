package panacea

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GrpcClient struct {
	conn *grpc.ClientConn
}

func NewGrpcClient(conf *config.Config) (*GrpcClient, error) {
	log.Infof("dialing to Panacea gRPC endpoint: %s", conf.Panacea.GRPCAddr)
	conn, err := grpc.Dial(conf.Panacea.GRPCAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Panacea: %w", err)
	}

	return &GrpcClient{
		conn: conn,
	}, nil
}

func (c *GrpcClient) Close() error {
	log.Info("closing Panacea gRPC connection")
	return c.conn.Close()
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
