package ipfs_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-doracle/ipfs"
	"github.com/medibloc/panacea-doracle/ipfs/suite"
	"github.com/stretchr/testify/require"
)

type ipfsTestSuite struct {
	suite.TestSuiteIpfs
}

func TestIpfs(t *testing.T) {
	suite.Run(t, new(ipfsTestSuite))
}

type testdata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

const (
	ipfsNodeAddr = "localhost:5001"
)

func (suite *ipfsTestSuite) TestIpfsAdd() {
	newIpfs := ipfs.NewIpfs(ipfsNodeAddr)

	testData := &testdata{
		Name:        "panacea",
		Description: "medibloc mainnet",
	}

	testDataBz, err := json.Marshal(testData)
	require.NoError(suite.T(), err)

	_, err = newIpfs.Add(testDataBz)
	require.NoError(suite.T(), err)
}

func (suite *ipfsTestSuite) TestIpfsGet() {
	newIpfs := ipfs.NewIpfs(ipfsNodeAddr)

	file, err := os.ReadFile("testdata/test_deal.json")
	require.NoError(suite.T(), err)

	cid, err := newIpfs.Add(file)
	require.NoError(suite.T(), err)

	getStrings, err := newIpfs.Get(cid)
	require.NoError(suite.T(), err)

	var deal types.Deal
	err = json.Unmarshal(file, &deal)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), deal.DataSchema, getStrings)
}
