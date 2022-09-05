package sgx_test

import (
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestSealAndUnsealFile(t *testing.T) {
	data := []byte("hello world")

	testDir := "/test"
	err := os.MkdirAll(testDir, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	dataPath := filepath.Join(testDir, "temp.sealed")

	err = sgx.SealToFile(data, dataPath)
	require.NoError(t, err)

	dataRetrieved, err := sgx.UnsealFromFile(dataPath)
	require.NoError(t, err)

	require.Equal(t, dataRetrieved, data)
}
