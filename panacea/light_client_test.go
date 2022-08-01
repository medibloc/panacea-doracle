package panacea

import (
	"context"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	chainID = "panacea-3"
	rpcAddr = "https://rpc.gopanacea.org:443"
)

var (
	ctx = context.Background()
)

func TestProof(t *testing.T) {

	// set initial trusted height and block hash
	height := 99
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")

	// create new light client
	lc, err := NewLightClient(ctx, chainID, rpcAddr, height, hash)
	require.NoError(t, err)

	//get latest trusted block for test
	trustedBlock, err := lc.Update(ctx, time.Now())

	// query getOracle
	var oracle types.Oracle
	var storeKey = "oracleStoreKey"
	var address sdk.AccAddress
	var key = types.GetOracleKey(address)

	err = GetStoreData(ctx, rpcAddr, lc, storeKey, key, trustedBlock.Height, &oracle)

}
