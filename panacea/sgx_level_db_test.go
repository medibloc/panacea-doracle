package panacea_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	sgxdb "github.com/medibloc/panacea-doracle/tm-db"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
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

	// get Block info using sgxLevelDB function
	storedLightBlock, err := lightClient.TrustedLightBlock(height)
	require.NoError(t, err)
	fmt.Println(storedLightBlock)

	// directly get data from DB
	dbDir, err := ioutil.TempDir("", "light-client")
	require.NoError(t, err)
	defer os.RemoveAll(dbDir)

	db, err := sgxdb.NewGoLevelDB("light-client-db", dbDir)
	require.NoError(t, err)
	getFromLevelDB, err := db.Db.Get([]byte(fmt.Sprintf("lb/%s/%20d", "panacea-3", height)), nil)
	require.NoError(t, err)

	fmt.Println(getFromLevelDB)

	//unseal data from levelDB
	unsealedBlock, err := sgx.Unseal(getFromLevelDB, true)
	require.NoError(t, err)
	fmt.Println(unsealedBlock)
}
