package panacea

import (
	"context"
	"encoding/hex"
	"fmt"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	chainID = "local"
	rpcAddr = "127.0.0.1:26657"
)

var (
	ctx = context.Background()
)

func GetAccount(address string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(address, "panacea")
}

func TestProof(t *testing.T) {

	// set initial trusted height and block hash (local panacea)
	height := 3
	hash, err := hex.DecodeString("2D34ACB2574FFCE1E47431D1C40D5468D27793E2D8227FBE855A2060836212FF")

	// create new light client
	lc, err := NewLightClient(ctx, chainID, rpcAddr, height, hash)
	require.NoError(t, err)

	//get latest trusted block for test
	trustedBlock, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)

	// query GetBalance
	// panacea13juzqmdmy7eh98awskqavwzge40h7c04aycduy

	var resultBalance = &types2.Any{}
	acc, err := GetAccount("panacea13juzqmdmy7eh98awskqavwzge40h7c04aycduy")
	require.NoError(t, err)
	key := authtypes.AddressStoreKey(acc)
	var storeKey = authtypes.StoreKey

	err = GetStoreData(ctx, "http://127.0.0.1:26657", lc, storeKey, key, trustedBlock.Height, resultBalance)
	require.NoError(t, err)
	fmt.Println(resultBalance)

}
