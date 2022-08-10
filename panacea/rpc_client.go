package panacea

import (
	"context"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	sgxdb "github.com/medibloc/panacea-doracle/tm-db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	"github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	"github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"os"
	"time"
)

const blockPeriod = 6 * time.Second

type RpcClient struct {
	RpcClient   *rpchttp.HTTP
	LightClient *light.Client
}

// NewRpcClient set RpcClient with trustedBlock Info and returns, if successful,
// a RpcClient that can be used to query data with light client verification & merkle proof.
// Period of trustOptions can be changed according to slashing period.
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

	//dbDir, err := ioutil.TempDir("", "light-client-example")
	//if err != nil {
	//	return nil, err
	//}
	//db, err := sgxdb.NewGoLevelDB("light-client-db", dbDir)

	if _, err := os.Stat("./light_client"); os.IsNotExist(err) {
		err = os.Mkdir("./light_client", 0700)
		if err != nil {
			return nil, err
		}
	}
	db, err := sgxdb.NewGoLevelDB("light-client-db", "./light_client")

	store := dbs.New(db, chainID)

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

// GetStoreData get data from panacea with storeKey and key, then verify queried data with light client and merkle proof.
// the returned data type is ResponseQuery.value ([]byte), so recommend to convert to expected type
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
	time.Sleep(blockPeriod) // AppHash for query is in the next block, so have to wait until the next block is confirmed

	// wait a creation of the next block
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

	merklePath := types.NewMerklePath(storeKey, string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}
