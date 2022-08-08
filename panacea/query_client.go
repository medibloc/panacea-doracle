package panacea

import (
	"context"
	"fmt"
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

func (c QueryClient) GetBalance(address string) (sdk.Coins, error) {
	acc, err := sdk.GetFromBech32(address, "panacea")
	if err != nil {
		return nil, err
	}

	var denom = "umed"
	key := banktypes.CreateAccountBalancesPrefix(acc)
	newkey := cloneAppend(key, []byte(denom))

	bz, err := c.rpcClient.GetStoreData(context.Background(), banktypes.StoreKey, newkey)
	if err != nil {
		return nil, err
	}
	fmt.Println("bz: ", bz)

	var balanceAny codectypes.Any
	err = balanceAny.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	var balance sdk.Coins
	err = c.interfaceRegistry.UnpackAny(&balanceAny, &balance)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (c QueryClient) GetTopic(address string, topicName string) (aoltypes.Topic, error) {
	acc, err := sdk.GetFromBech32(address, "panacea")
	if err != nil {
		return aoltypes.Topic{}, err
	}

	key := aoltypes.TopicCompositeKey{OwnerAddress: acc, TopicName: topicName}
	topicKey := compkey.MustEncode(&key)
	bz, err := c.rpcClient.GetStoreData(context.Background(), aoltypes.StoreKey, topicKey)

	var topicAny codectypes.Any
	err = topicAny.Unmarshal(bz)
	if err != nil {
		return aoltypes.Topic{}, err
	}

	var topic aoltypes.Topic
	err = c.interfaceRegistry.UnpackAny(&topicAny, &topic)
	if err != nil {
		return aoltypes.Topic{}, err
	}

	return topic, nil
}

func cloneAppend(bz []byte, tail []byte) (res []byte) {
	res = make([]byte, len(bz)+len(tail))
	copy(res, bz)
	copy(res[len(bz):], tail)
	return
}
