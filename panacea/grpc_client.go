package panacea

import (
	"context"
	"fmt"
	"net/url"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/medibloc/panacea-doracle/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcClient struct {
	conn *grpc.ClientConn
}

func NewGrpcClient(conf *config.Config) (*GrpcClient, error) {
	log.Infof("dialing to Panacea gRPC endpoint: %s", conf.Panacea.GRPCAddr)

	//var targetUrl string
	parsedUrl, err := url.Parse(conf.Panacea.GRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gRPC endpoint. please use absolute URL (scheme://host:port): %w", err)
	}

	var cred grpc.DialOption

	if parsedUrl.Scheme == "https" {
		tlsConfig, err := config.CreateTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS cofig: %w", err)
		}
		cred = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
		log.Infof("tls connection! %v", parsedUrl.Host)
	} else {
		cred = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(parsedUrl.Host, cred)
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
