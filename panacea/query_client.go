package panacea

import (
	"context"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/medibloc/panacea-core/v2/types/compkey"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
)

type QueryClient struct {
	rpcClient         *RpcClient
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewQueryClient set QueryClient with rpcClient & and returns, if successful,
// a QueryClient that can be used to add query function.
func NewQueryClient(ctx context.Context, chainID, rpcAddr string, trustedBlockHeight int, trustedBlockHash []byte) (*QueryClient, error) {
	rpcClient, err := NewRpcClient(ctx, chainID, rpcAddr, trustedBlockHeight, trustedBlockHash)
	if err != nil {
		return nil, err
	}

	return &QueryClient{
		rpcClient:         rpcClient,
		interfaceRegistry: makeInterfaceRegistry(),
	}, nil
}

// Below are examples of query function that use GetStoreData function to verify queried result.
// Need to set storeKey and key inside the query function, and change type to expected type.

// GetAccount returns account from address.
func (c QueryClient) GetAccount(address string) (authtypes.AccountI, error) {
	acc, err := sdk.GetFromBech32(address, "panacea")
	if err != nil {
		return nil, err
	}

	key := authtypes.AddressStoreKey(acc)
	bz, err := c.rpcClient.GetStoreData(context.Background(), authtypes.StoreKey, key)
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
	acc, err := sdk.GetFromBech32(address, "panacea")
	if err != nil {
		return sdk.Coin{}, err
	}

	var denom = "umed"
	key := append(banktypes.BalancesPrefix, append(acc, []byte(denom)...)...)

	bz, err := c.rpcClient.GetStoreData(context.Background(), banktypes.StoreKey, key)
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
	acc, err := sdk.GetFromBech32(address, "panacea")
	if err != nil {
		return aoltypes.Topic{}, err
	}

	key := aoltypes.TopicCompositeKey{OwnerAddress: acc, TopicName: topicName}
	topicKey := append(aoltypes.TopicKeyPrefix, compkey.MustEncode(&key)...)
	bz, err := c.rpcClient.GetStoreData(context.Background(), aoltypes.StoreKey, topicKey)

	var topic aoltypes.Topic
	err = topic.Unmarshal(bz)
	if err != nil {
		return aoltypes.Topic{}, err
	}

	return topic, nil
}
