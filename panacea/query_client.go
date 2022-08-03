package panacea

import (
	"context"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
