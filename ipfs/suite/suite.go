package suite

import (
	"testing"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type TestSuiteIpfs struct {
	suite.Suite

	dktPool     *dockertest.Pool
	dktResource *dockertest.Resource
}

func Run(t *testing.T, s suite.TestingSuite) {
	suite.Run(t, s)
}

func (suite *TestSuiteIpfs) SetupSuite() {
	dktPool, err := dockertest.NewPool("")
	if err != nil {
		log.Panicf("Could not connect to docker: %v", err)
	}
	suite.dktPool = dktPool
}

func (suite *TestSuiteIpfs) SetupTest() {
	var err error

	suite.dktResource, err = suite.dktPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:         "ipfs_host",
			Repository:   "ipfs/kubo",
			Tag:          "latest",
			ExposedPorts: []string{"4001", "5001", "8080"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true // so that stopped containers are removed automatically
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		log.Panicf("Could not start resource: %v", err)
	}

	if err := suite.dktPool.Retry(func() error {
		_ = shell.NewShell("localhost:5001")
		return nil
	}); err != nil {
		log.Panicf("Could not connect to ipfs: %s", err)
	}
}

func (suite *TestSuiteIpfs) TearDownTest() {
	if err := suite.dktPool.Purge(suite.dktResource); err != nil {
		log.Printf("Could not purge resource: %s", err)
	}
}
