package panacea

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
	"github.com/medibloc/panacea-doracle/integration/rest"
	"github.com/medibloc/panacea-doracle/integration/suite"
	dbm "github.com/tendermint/tm-db"

	"github.com/medibloc/panacea-doracle/config"
	"github.com/stretchr/testify/require"
)

type queryClientTestSuite struct {
	suite.TestSuite

	chainID           string
	validatorMnemonic string
}

func TestQueryClient(t *testing.T) {
	initScriptPath, err := filepath.Abs("testdata/panacea-core-init.sh")
	require.NoError(t, err)

	chainID := "testing"
	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	validatorMnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	suite.Run(t, &queryClientTestSuite{
		suite.NewTestSuite(
			initScriptPath,
			[]string{
				fmt.Sprintf("CHAIN_ID=%s", chainID),
				fmt.Sprintf("MNEMONIC=%s", validatorMnemonic),
			},
		),
		chainID,
		validatorMnemonic,
	})
}

func (suite *queryClientTestSuite) TestGetAccount() {
	trustedBlockInfo, conf := suite.prepare()

	queryClient, err := NewQueryClientWithDB(context.Background(), conf, trustedBlockInfo, dbm.NewMemDB())
	require.NoError(suite.T(), err)
	defer queryClient.Close()

	var wg sync.WaitGroup
	accAddr := suite.AccAddressFromMnemonic(suite.validatorMnemonic, 0, 0)

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

	queryClient, err := NewQueryClientWithDB(context.Background(), conf, trustedBlockInfo, db)
	require.NoError(suite.T(), err)

	lastTrustedHeight, err := queryClient.lightClient.LastTrustedHeight()
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), lastTrustedHeight, trustedBlockInfo.TrustedBlockHeight)

	err = queryClient.Close() // here, memdb is not closed because MemDB.Close() is actually empty
	require.NoError(suite.T(), err)

	// try to load query client, instead of creating it
	queryClient, err = NewQueryClientWithDB(context.Background(), conf, nil, db)
	require.NoError(suite.T(), err)

	lastTrustedHeight2, err := queryClient.lightClient.LastTrustedHeight()
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), lastTrustedHeight2, lastTrustedHeight)
}

func (suite *queryClientTestSuite) TestGetOracleUpgradeInfoEmptyValue() {
	trustedBlockInfo, conf := suite.prepare()

	queryClient, err := NewQueryClientWithDB(context.Background(), conf, trustedBlockInfo, dbm.NewMemDB())
	require.NoError(suite.T(), err)
	defer queryClient.Close()

	upgradeInfo, err := queryClient.GetOracleUpgradeInfo()
	require.Nil(suite.T(), upgradeInfo)
	require.ErrorIs(suite.T(), err, ErrEmptyValue)
}

func (suite *queryClientTestSuite) prepare() (*TrustedBlockInfo, *config.Config) {
	hash, height, err := rest.QueryLatestBlock(suite.PanaceaEndpoint("http", 1317))
	require.NoError(suite.T(), err)

	trustedBlockInfo := &TrustedBlockInfo{
		TrustedBlockHeight: height,
		TrustedBlockHash:   hash,
	}

	conf := &config.Config{
		Panacea: config.PanaceaConfig{
			GRPCAddr:                suite.PanaceaEndpoint("tcp", 9090),
			RPCAddr:                 suite.PanaceaEndpoint("tcp", 26657),
			ChainID:                 suite.chainID,
			LightClientPrimaryAddr:  suite.PanaceaEndpoint("tcp", 26657),
			LightClientWitnessAddrs: []string{suite.PanaceaEndpoint("tcp", 26657)},
		},
	}

	return trustedBlockInfo, conf
}
