package event

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/ipfs"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
)

// Reactor contains all ingredients needed for handling this type of event
type Reactor interface {
	GRPCClient() *panacea.GrpcClient
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() *panacea.QueryClient
	Ipfs() *ipfs.Ipfs
}
