package panacea

import (
	"context"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	"github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	"github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	dbm "github.com/tendermint/tm-db"
	"time"
)

type RpcClient struct {
	RpcClient   *rpchttp.HTTP
	LightClient *light.Client
}

func NewRpcClient(ctx context.Context, chainID, rpcAddr string, trustedHeight int, trustedBlockHash []byte) (*RpcClient, error) {
	rpcClient, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	trustOptions := light.TrustOptions{
		Period: 365 * 24 * time.Hour,
		Height: int64(trustedHeight),
		Hash:   trustedBlockHash,
	}

	pv, err := http.New(chainID, rpcAddr)
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

	return &RpcClient{
		RpcClient:   rpcClient,
		LightClient: lc,
	}, nil
}

func (q RpcClient) GetStoreData(ctx context.Context, storeKey string, key []byte) ([]byte, error) {
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
	time.Sleep(blockPeriod) // AppHash for query is in the next block, so have to wait until next block is confirmed

	// wait a creation next block
	textTrustedBlock, err := q.LightClient.VerifyLightBlockAtHeight(ctx, trustedBlock.Height+1, time.Now())
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
	merkleRootKey := types.NewMerkleRoot(textTrustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath(authtypes.StoreKey, string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}
