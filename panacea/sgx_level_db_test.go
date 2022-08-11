package panacea_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	sgxdb "github.com/medibloc/panacea-doracle/tm-db"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSgxLevelDB(t *testing.T) {
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	queryClient, err := panacea.NewQueryClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)

	require.NoError(t, err)

	lightClient := queryClient.RpcClient.LightClient
	lightClient.VerifyLightBlockAtHeight(ctx, 1000, time.Now())

	// get Block info using sgxLevelDB function
	storedLightBlock, err := lightClient.TrustedLightBlock(1000)
	require.NoError(t, err)
	fmt.Println(storedLightBlock)

	// directly get data from DB
	db, err := sgxdb.NewGoLevelDB("light-client-db", "../data")
	require.NoError(t, err)
	getFromLevelDB, err := db.Db.Get([]byte(fmt.Sprintf("lb/%s/%20d", "panacea-3", 1000)), nil)
	require.NoError(t, err)

	fmt.Println(getFromLevelDB)

	//unseal data from levelDB
	unsealedBlock, err := sgx.Unseal(getFromLevelDB, true)
	require.NoError(t, err)
	fmt.Println(unsealedBlock)
}
