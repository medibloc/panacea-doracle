package panacea_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	"testing"
	"time"
)

func TestSgxLevelDB(t *testing.T) {
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	queryClient, err := panacea.NewQueryClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)
	require.NoError(t, err)

	lightClient := queryClient.LightClient
	_, err = lightClient.VerifyLightBlockAtHeight(ctx, 1000, time.Now())
	require.NoError(t, err)

	// get Block info using sgxLevelDB function
	storedLightBlock, err := lightClient.TrustedLightBlock(1000)
	require.NoError(t, err)
	fmt.Println("storedLightBlock at ", storedLightBlock.Height, " : ", storedLightBlock.Hash())

	// directly get data from DB
	db := queryClient.Db
	bz, err := db.Db.Get([]byte(fmt.Sprintf("lb/%s/%020d", "panacea-3", 1000)), nil)
	require.NoError(t, err)

	// unseal data
	unsealedBz, err := sgx.Unseal(bz, true)
	require.NoError(t, err)

	var lbpb tmproto.LightBlock
	err = lbpb.Unmarshal(unsealedBz)
	require.NoError(t, err)

	lightBlock, err := types.LightBlockFromProto(&lbpb)
	require.NoError(t, err)

	fmt.Println("GetFromLevelDB: ", lightBlock)
}
