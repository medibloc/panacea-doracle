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

type registerOracleEventTestSuite struct {
	suite.TestSuite

	chainID           string
	validatorMnemonic string
}

func TestRegisterOracleEvent(t *testing.T) {
	initScriptPath, err := filepath.Abs("../testdata/panacea-core-init.sh")
	require.NoError(t, err)

	chainID := "testing"
	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)
	validatorMnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	suite.Run(t, &registerOracleEventTestSuite{
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

func (suite *registerOracleEventTestSuite) TestVerifyAndGetVoteOptionInvalidTrustedBlockHash() {
	trustedBlockInfo, conf := suite.prepare()

	svc, err := service.NewTestServiceWithoutSGX(conf, trustedBlockInfo)
	require.NoError(suite.T(), err)

	oracleRegistration := &oracletypes.OracleRegistration{
		TrustedBlockHeight: trustedBlockInfo.TrustedBlockHeight,
		TrustedBlockHash:   []byte("invalid"),
	}

	e := NewRegisterOracleEvent(svc)
	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)

	require.ErrorContains(suite.T(), err, "failed to verify trusted block information")
	require.Equal(suite.T(), oracletypes.VOTE_OPTION_NO, voteOption)
}

func (suite *registerOracleEventTestSuite) TestVerifyAndGetVoteOptionHigherTrustedBlockHeight() {
	trustedBlockInfo, conf := suite.prepare()

	svc, err := service.NewTestServiceWithoutSGX(conf, trustedBlockInfo)
	require.NoError(suite.T(), err)

	oracleRegistration := &oracletypes.OracleRegistration{
		TrustedBlockHeight: 100,
		TrustedBlockHash:   trustedBlockInfo.TrustedBlockHash,
	}

	e := NewRegisterOracleEvent(svc)
	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)

	require.ErrorContains(suite.T(), err, "not found light block.")
	require.Equal(suite.T(), oracletypes.VOTE_OPTION_NO, voteOption)
}

func (suite *registerOracleEventTestSuite) prepare() (*panacea.TrustedBlockInfo, *config.Config) {
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
