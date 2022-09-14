package panacea

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	dbm "github.com/tendermint/tm-db"

	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/integration"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type queryClientTestSuite struct {
	integration.TestSuite
}

func TestQueryClient(t *testing.T) {
	absPath, err := filepath.Abs("testdata/panacea-core-init.sh")
	require.NoError(t, err)

	suite.Run(t, &queryClientTestSuite{integration.NewTestSuite(absPath)})
}

func (suite *queryClientTestSuite) TestGetAccount() {
	trustedBlockInfo, conf := suite.prepare()

	queryClient, err := newQueryClientWithDB(context.Background(), conf, trustedBlockInfo, dbm.NewMemDB())
	require.NoError(suite.T(), err)
	defer queryClient.Close()

	var wg sync.WaitGroup
	accAddr := suite.ValidatorAccAddress()

	for i := 0; i < 10; i++ { // to check if queryClient is goroutine-safe
		wg.Add(1)

		go func() {
			defer wg.Done()

			acc, err := queryClient.GetAccount(accAddr)
			require.NoError(suite.T(), err)

			address, err := bech32.ConvertAndEncode("panacea", acc.GetPubKey().Address().Bytes())
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), accAddr, address)
		}()
	}

	wg.Wait()
}

func (suite *queryClientTestSuite) TestLoadQueryClient() {
	trustedBlockInfo, conf := suite.prepare()

	db := dbm.NewMemDB()

	queryClient, err := newQueryClientWithDB(context.Background(), conf, trustedBlockInfo, db)
	require.NoError(suite.T(), err)

	lastTrustedHeight, err := queryClient.lightClient.LastTrustedHeight()
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), lastTrustedHeight, trustedBlockInfo.TrustedBlockHeight)

	err = queryClient.Close() // here, memdb is not closed because MemDB.Close() is actually empty
	require.NoError(suite.T(), err)

	// try to load query client, instead of creating it
	queryClient, err = newQueryClientWithDB(context.Background(), conf, nil, db)
	require.NoError(suite.T(), err)

	lastTrustedHeight2, err := queryClient.lightClient.LastTrustedHeight()
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), lastTrustedHeight2, lastTrustedHeight)
}

func (suite *queryClientTestSuite) prepare() (*TrustedBlockInfo, *config.Config) {
	hash, height, err := integration.QueryLatestBlock(suite.PanaceaEndpoint("http", 1317))
	require.NoError(suite.T(), err)

	trustedBlockInfo := &TrustedBlockInfo{
		TrustedBlockHeight: height,
		TrustedBlockHash:   hash,
	}

	conf := &config.Config{
		Panacea: config.PanaceaConfig{
			GRPCAddr:                suite.PanaceaEndpoint("tcp", 9090),
			RPCAddr:                 suite.PanaceaEndpoint("tcp", 26657),
			ChainID:                 suite.ChainID,
			LightClientPrimaryAddr:  suite.PanaceaEndpoint("tcp", 26657),
			LightClientWitnessAddrs: []string{suite.PanaceaEndpoint("tcp", 26657)},
		},
	}

	return trustedBlockInfo, conf
}
