package oracle

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cosmos/go-bip39"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/integration/rest"
	"github.com/medibloc/panacea-doracle/integration/service"
	"github.com/medibloc/panacea-doracle/integration/suite"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
)

var _ event.Reactor = (*service.TestServiceWithoutSGX)(nil)

type upgradeOracleEventTestSuite struct {
	suite.TestSuite

	chainID           string
	validatorMnemonic string
	uniqueID          string
	upgradeUniqueID   string
}

func TestUpgradeOracleEvent(t *testing.T) {
	initScriptPath, err := filepath.Abs("../testdata/panacea-core-init.sh")
	require.NoError(t, err)

	chainID := "testing"
	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	validatorMnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)
	uniqueID := "uniqueID"
	upgradeUniqueID := "upgradeUniqueID"

	suite.Run(t, &upgradeOracleEventTestSuite{
		suite.NewTestSuite(
			initScriptPath,
			[]string{
				fmt.Sprintf("CHAIN_ID=%s", chainID),
				fmt.Sprintf("MNEMONIC=%s", validatorMnemonic),
				fmt.Sprintf("UNIQUE_ID=%s", uniqueID),
				fmt.Sprintf("UPGRADE_UNIQUE_ID=%s", upgradeUniqueID),
			},
		),
		chainID,
		validatorMnemonic,
		uniqueID,
		upgradeUniqueID,
	})
}

func (suite *upgradeOracleEventTestSuite) TestVerifyAndGetVoteOptionNotFoundOracleUpgradeInfo() {
	trustedBlockInfo, conf := suite.prepare()

	svc, err := service.NewTestServiceWithoutSGX(conf, trustedBlockInfo)
	require.NoError(suite.T(), err)

	oracleRegistration := &oracletypes.OracleRegistration{
		UniqueId:           suite.upgradeUniqueID,
		TrustedBlockHeight: trustedBlockInfo.TrustedBlockHeight,
		TrustedBlockHash:   trustedBlockInfo.TrustedBlockHash,
	}

	e := NewUpgradeOracleEvent(svc)
	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)

	require.ErrorContains(suite.T(), err, "not found oracle upgrade info.")
	require.Equal(suite.T(), oracletypes.VOTE_OPTION_NO, voteOption)
}

func (suite *upgradeOracleEventTestSuite) prepare() (*panacea.TrustedBlockInfo, *config.Config) {
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
