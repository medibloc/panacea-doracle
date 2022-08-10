package panacea_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSgxLevelDB(t *testing.T) {
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	var height int64 = 1000

	require.NoError(t, err)
	ctx := context.Background()

	rpcClient, err := panacea.NewRpcClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)
	require.NoError(t, err)

	lightClient := rpcClient.LightClient
	lightClient.VerifyLightBlockAtHeight(ctx, height, time.Now())

	storedLightBlock, err := lightClient.TrustedLightBlock(height)
	require.NoError(t, err)

	fmt.Println(storedLightBlock)

	fmt.Println("sealed hash: ", storedLightBlock.Hash())
	blockHash, err := hex.DecodeString("6DD94FFAFC97EBDC0A5A6A32F532B69D266B6C6597A10E126B90FCE49032FC5C")
	fmt.Println("block hash: ", blockHash)
	unsealedHash, err := sgx.Unseal(storedLightBlock.Hash(), true)
	require.NoError(t, err)
	fmt.Println("unsealed hash: ", unsealedHash)

}
