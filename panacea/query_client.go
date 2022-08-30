package panacea

import (
	"context"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	tmhttp "github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	"github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	dbm "github.com/tendermint/tm-db"
	"time"
)

const (
	blockPeriod = 6 * time.Second
)

type TrustedBlockInfo struct {
	TrustedBlockHeight int64
	TrustedBlockHash   []byte
}

type QueryClient struct {
	RpcClient         *rpchttp.HTTP
	LightClient       *light.Client
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewQueryClient set QueryClient with rpcClient & and returns, if successful,
// a QueryClient that can be used to add query function.
func NewQueryClient(ctx context.Context, chainID, rpcAddr string, trustedBlockHeight int, trustedBlockHash []byte) (*QueryClient, error) {
	rpcClient, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	trustOptions := light.TrustOptions{
		Period: 2 * 365 * 24 * time.Hour,
		Height: int64(trustedBlockHeight),
		Hash:   trustedBlockHash,
	}

	pv, err := tmhttp.New(chainID, rpcAddr)
	if err != nil {
		return nil, err
	}
	pvs := []provider.Provider{pv}
	store := dbs.New(dbm.NewMemDB(), chainID)

	lc, err := light.NewClient(
		ctx,
		chainID,
		trustOptions,
		pv,
		pvs,
		store,
		light.SkippingVerification(light.DefaultTrustLevel),
		light.Logger(log.TestingLogger()),
	)
	if err != nil {
		return nil, err
	}

	return &QueryClient{
		RpcClient:         rpcClient,
		LightClient:       lc,
		interfaceRegistry: makeInterfaceRegistry(),
	}, nil
}

// GetStoreData get data from panacea with storeKey and key, then verify queried data with light client and merkle proof.
// the returned data type is ResponseQuery.value ([]byte), so recommend to convert to expected type
func (q QueryClient) GetStoreData(ctx context.Context, storeKey string, key []byte) ([]byte, error) {
	trustedBlock, err := q.LightClient.Update(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	//set queryOption prove to true
	option := client.ABCIQueryOptions{
		Prove:  true,
		Height: trustedBlock.Height,
	}
	// query to kv store with proof option
	result, err := q.RpcClient.ABCIQueryWithOptions(ctx, fmt.Sprintf("/store/%s/key", storeKey), key, option)
	if err != nil {
		return nil, err
	}

	// get trustedBlock at blockHeight+1
	time.Sleep(blockPeriod) // AppHash for query is in the next block, so have to wait until the next block is confirmed

	// wait a creation of the next block
	nextTrustedBlock, err := q.LightClient.VerifyLightBlockAtHeight(ctx, trustedBlock.Height+1, time.Now())
	if err != nil {
		return nil, err
	}

	// verify query result with merkle proof & trusted block info
	proofOps := result.Response.ProofOps
	merkleProof, err := types.ConvertProofs(proofOps)
	if err != nil {
		return nil, err
	}

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(nextTrustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath(storeKey, string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

// Below are examples of query function that use GetStoreData function to verify queried result.
// Need to set storeKey and key inside the query function, and change type to expected type.

// GetAccount returns account from address.
func (q QueryClient) GetAccount(address string) (authtypes.AccountI, error) {
	acc, err := GetAccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	key := authtypes.AddressStoreKey(acc)
	bz, err := q.GetStoreData(context.Background(), authtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var accountAny codectypes.Any
	err = accountAny.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = q.interfaceRegistry.UnpackAny(&accountAny, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (q QueryClient) GetOracleRegistration(oracleAddr, uniqueID string) (*oracletypes.OracleRegistration, error) {

	acc, err := GetAccAddressFromBech32(oracleAddr)
	if err != nil {
		return nil, err
	}

	key := oracletypes.GetOracleRegistrationKey(uniqueID, acc)

	bz, err := q.GetStoreData(context.Background(), oracletypes.StoreKey, key)
	if err != nil {
		return nil, err
	}
	var oracleRegistration oracletypes.OracleRegistration
	err = oracleRegistration.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	return &oracleRegistration, nil
}
