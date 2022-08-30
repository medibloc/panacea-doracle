package panacea

import (
	"context"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	"github.com/medibloc/panacea-core/v2/types/compkey"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	"github.com/sirupsen/logrus"
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
	denom       = "umed"
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
	// call refresh every minute
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			if err := refresh(ctx, lc, trustOptions.Period); err != nil {
			    logrus.Errorf("light client refresh error: %v", err)
			}
		}
	}()

	return &QueryClient{
		RpcClient:         rpcClient,
		LightClient:       lc,
		interfaceRegistry: makeInterfaceRegistry(),
	}, nil
}

// refresh update light block, when the last light block has been updated more than trustPeriod * 2/3.
func refresh(ctx context.Context, lc *light.Client, trustPeriod time.Duration) error {
	logrus.Info("check latest light block")
	lastBlockHeight, err := lc.LastTrustedHeight()
	if err != nil {
		return err
	}
	lastBlock, err := lc.TrustedLightBlock(lastBlockHeight)
	if err != nil {
		return err
	}
	lastBlockTime := lastBlock.Time
	currentTime := time.Now()
	timeDiff := currentTime.Sub(lastBlockTime)
	if timeDiff > trustPeriod*2/3 {
		logrus.Info("update latest light block")
		_, err = lc.Update(ctx, time.Now())
		if err != nil {
			return err
		}
	}
	return nil
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

	// get nextTrustedBlock from LightClient Primary.
	// It returns a new light block on a successful update. Otherwise, it returns nil
	nextTrustedBlock, err := q.LightClient.Update(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	// if nextTrustedBlock is not a new block, wait until the next block is confirmed
	// AppHash for query is in the next block, so have to get nextTrustedBlock to verify query
	if nextTrustedBlock == nil {
		// wait a creation of the next block
		time.Sleep(blockPeriod)

		// get trustedBlock at blockHeight+1
		nextTrustedBlock, err = q.LightClient.VerifyLightBlockAtHeight(ctx, trustedBlock.Height+1, time.Now())
		if err != nil {
			return nil, err
		}
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
func (c QueryClient) GetAccount(address string) (authtypes.AccountI, error) {
	acc, err := GetAccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	key := authtypes.AddressStoreKey(acc)
	bz, err := c.GetStoreData(context.Background(), authtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var accountAny codectypes.Any
	err = accountAny.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = c.interfaceRegistry.UnpackAny(&accountAny, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetBalance returns balance from address.
func (c QueryClient) GetBalance(address string) (sdk.Coin, error) {
	acc, err := GetAccAddressFromBech32(address)
	if err != nil {
		return sdk.Coin{}, err
	}

	key := append(banktypes.BalancesPrefix, append(acc, []byte(denom)...)...)

	bz, err := c.GetStoreData(context.Background(), banktypes.StoreKey, key)
	if err != nil {
		return sdk.Coin{}, err
	}

	var balance sdk.Coin
	err = balance.Unmarshal(bz)
	if err != nil {
		return sdk.Coin{}, err
	}

	return balance, nil
}

// GetTopic returns topic from address and topicName.
func (c QueryClient) GetTopic(address string, topicName string) (aoltypes.Topic, error) {
	acc, err := GetAccAddressFromBech32(address)
	if err != nil {
		return aoltypes.Topic{}, err
	}

	key := aoltypes.TopicCompositeKey{OwnerAddress: acc, TopicName: topicName}
	topicKey := append(aoltypes.TopicKeyPrefix, compkey.MustEncode(&key)...)
	bz, err := c.GetStoreData(context.Background(), aoltypes.StoreKey, topicKey)
	if err != nil {
		return aoltypes.Topic{}, err
	}

	var topic aoltypes.Topic
	err = topic.Unmarshal(bz)
	if err != nil {
		return aoltypes.Topic{}, err
	}

	return topic, nil
}
