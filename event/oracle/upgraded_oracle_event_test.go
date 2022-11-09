package oracle

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	datadealevent "github.com/medibloc/panacea-doracle/event/datadeal"
	"github.com/medibloc/panacea-doracle/integration/rest"
	"github.com/medibloc/panacea-doracle/integration/service"
	"github.com/medibloc/panacea-doracle/integration/suite"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
)

type upgradedOracleEventTestSuite struct {
	suite.TestSuite

	chainID           string
	validatorMnemonic string
	uniqueID          string
}

func TestUpgradedOracleEvent(t *testing.T) {
	initScriptPath, err := filepath.Abs("../testdata/panacea-core-init.sh")
	require.NoError(t, err)

	chainID := "testing"
	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	validatorMnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)
	uniqueID := hex.EncodeToString([]byte("uniqueID"))

	suite.Run(t, &upgradedOracleEventTestSuite{
		suite.NewTestSuite(
			initScriptPath,
			[]string{
				fmt.Sprintf("CHAIN_ID=%s", chainID),
				fmt.Sprintf("MNEMONIC=%s", validatorMnemonic),
				fmt.Sprintf("UNIQUE_ID=%s", uniqueID),
			},
		),
		chainID,
		validatorMnemonic,
		uniqueID,
	})
}

func (suite *upgradedOracleEventTestSuite) TestSameUniqueID() {
	trustedBlockInfo, conf := suite.prepare()

	enclaveInfo := sgx.NewEnclaveInfo(
		[]byte("productID"),
		[]byte("signerID"),
		[]byte("uniqueID"),
	)

	svc, err := service.NewTestService(conf, trustedBlockInfo, enclaveInfo)
	require.NoError(suite.T(), err)

	voteEvents := []event.Event{
		NewRegisterOracleEvent(svc),
		NewUpgradeOracleEvent(svc),
		datadealevent.NewDataVerificationEvent(svc),
		datadealevent.NewDataDeliveryVoteEvent(svc),
	}

	e := NewUpgradedOracleEvent(svc, voteEvents)
	err = e.setEnableVoteEvents()
	require.NoError(suite.T(), err)
	for _, voteEvent := range voteEvents {
		require.True(suite.T(), voteEvent.Enabled())
	}
}

func (suite *upgradedOracleEventTestSuite) TestNotSameUniqueID() {
	trustedBlockInfo, conf := suite.prepare()

	enclaveInfo := sgx.NewEnclaveInfo(
		[]byte("productID"),
		[]byte("signerID"),
		[]byte("upgradeUniqueID"),
	)

	svc, err := service.NewTestService(conf, trustedBlockInfo, enclaveInfo)
	require.NoError(suite.T(), err)

	voteEvents := []event.Event{
		NewRegisterOracleEvent(svc),
		NewUpgradeOracleEvent(svc),
		datadealevent.NewDataVerificationEvent(svc),
		datadealevent.NewDataDeliveryVoteEvent(svc),
	}

	e := NewUpgradedOracleEvent(svc, voteEvents)
	err = e.setEnableVoteEvents()
	require.NoError(suite.T(), err)
	for _, voteEvent := range voteEvents {
		require.False(suite.T(), voteEvent.Enabled())
	}
}

func (suite *upgradedOracleEventTestSuite) prepare() (*panacea.TrustedBlockInfo, *config.Config) {
	hash, height, err := rest.QueryLatestBlock(suite.PanaceaEndpoint("http", 1317))
	require.NoError(suite.T(), err)

	trustedBlockInfo := &panacea.TrustedBlockInfo{
		TrustedBlockHeight: height,
		TrustedBlockHash:   hash,
	}

	conf := &config.Config{
		BaseConfig: config.BaseConfig{
			OracleMnemonic: suite.validatorMnemonic,
			OracleAccNum:   0,
			OracleAccIndex: 0,
		},
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
