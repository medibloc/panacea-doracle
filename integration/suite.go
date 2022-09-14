package integration

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/panacea-doracle/crypto"
	"log"
	"path/filepath"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	ChainID                   string
	validatorMnemonic         string
	panaceaInitScriptDir      string
	panaceaInitScriptFilename string

	dktPool     *dockertest.Pool
	dktResource *dockertest.Resource
}

func NewTestSuite(panaceaInitScriptAbsPath, validatorMnemonic string) TestSuite {
	if !filepath.IsAbs(panaceaInitScriptAbsPath) {
		log.Panicf("path must be absolute: %s", panaceaInitScriptAbsPath)
	}
	dir, filename := filepath.Split(panaceaInitScriptAbsPath)

	if validatorMnemonic == "" {
		log.Panic("validatorMnemonic is empty")
	}

	return TestSuite{
		ChainID:                   "testing",
		validatorMnemonic:         validatorMnemonic,
		panaceaInitScriptDir:      dir,
		panaceaInitScriptFilename: filename,
	}
}

func (suite *TestSuite) SetupSuite() {
	dktPool, err := dockertest.NewPool("")
	if err != nil {
		log.Panicf("Could not connect to docker: %v", err)
	}
	suite.dktPool = dktPool
}

func (suite *TestSuite) SetupTest() {
	var err error

	suite.dktResource, err = suite.dktPool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "ghcr.io/medibloc/panacea-core",
			Tag:        "master",
			Cmd:        []string{"bash", fmt.Sprintf("/scripts/%s", suite.panaceaInitScriptFilename)},
			Env: []string{
				fmt.Sprintf("CHAIN_ID=%s", suite.ChainID),
				fmt.Sprintf("MNEMONIC=%s", suite.validatorMnemonic),
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true // so that stopped containers are removed automatically
			config.Mounts = []docker.HostMount{
				{
					Source: suite.panaceaInitScriptDir,
					Target: "/scripts",
					Type:   "bind",
				},
			}
		},
	)
	if err != nil {
		log.Panicf("Could not start resource: %v", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	suite.dktPool.MaxWait = 1 * time.Minute
	if err := suite.dktPool.Retry(func() error {
		_, height, err := QueryLatestBlock(suite.PanaceaEndpoint("http", 1317))
		if err != nil {
			return err
		} else if height < 2 {
			return fmt.Errorf("wait until the height >= 2 is produced") // so that light client proof can work
		}

		return nil
	}); err != nil {
		log.Panicf("Could not connect to panacea-core: %s", err)
	}
}

func (suite *TestSuite) TearDownTest() {
	if err := suite.dktPool.Purge(suite.dktResource); err != nil {
		log.Printf("Could not purge resource: %s", err)
	}
}

func (suite *TestSuite) PanaceaEndpoint(scheme string, port int) string {
	return fmt.Sprintf(
		"%s://localhost:%s",
		scheme,
		suite.dktResource.GetPort(fmt.Sprintf("%d/tcp", port)),
	)
}

func (suite *TestSuite) ValidatorAccAddress() string {
	key, err := crypto.GeneratePrivateKeyFromMnemonic(suite.validatorMnemonic, 371, 0, 0)
	if err != nil {
		log.Panic(err)
	}

	addr, err := bech32.ConvertAndEncode("panacea", key.PubKey().Address().Bytes())
	if err != nil {
		log.Panic(err)
	}
	return addr
}
