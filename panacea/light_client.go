package panacea

import (
	"context"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	"github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	"github.com/tendermint/tendermint/rpc/client"
	httprpc "github.com/tendermint/tendermint/rpc/client/http"

	dbm "github.com/tendermint/tm-db"
	"time"
)

const blockPeriod = 5 * time.Second

func NewLightClient(ctx context.Context, chainID string, rpcAddr string, trustedHeight int, trustedBlockHash []byte) (c *light.Client, err error) {
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
	store := dbs.New(dbm.NewMemDB())

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

	return lc, nil
}

func GetStoreData(ctx context.Context, rpcAddr string, lc *light.Client, storeKey string, key bytes.HexBytes, blockHeight int64, ptr codec.ProtoMarshaler) error {

	// connect to rpc client
	rpcClient, err := httprpc.New(rpcAddr)
	if err != nil {
		return err
	}

	//set queryOption prove to true
	option := client.ABCIQueryOptions{
		Prove:  true,
		Height: blockHeight,
	}
	// query to kv store with proof option
	result, err := rpcClient.ABCIQueryWithOptions(ctx, fmt.Sprintf("/store/%s/key", storeKey), key, option)
	if err != nil {
		return err
	}

	// get trustedBlock at blockHeight+1
	time.Sleep(blockPeriod) // AppHash for query is in the next block, so have to wait until next block is confirmed
	trustedBlock, err := lc.VerifyLightBlockAtHeight(ctx, blockHeight+1, time.Now())
	if err != nil {
		return err
	}

	// verify query result with merkle proof & trusted block info
	proofOps := result.Response.ProofOps
	merkleProof, err := types.ConvertProofs(proofOps)
	if err != nil {
		return err
	}

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(trustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath(authtypes.StoreKey, string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	if err != nil {
		return err
	}

	// convert result to expected data type
	err = ptr.Unmarshal(result.Response.Value)
	if err != nil {
		return err
	}

	return nil
}
