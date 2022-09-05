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

	dir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	dataPath := filepath.Join(dir, "temp.sealed")

	err = sgx.SealToFile(data, dataPath)
	require.NoError(t, err)

	dataRetrieved, err := sgx.UnsealFromFile(dataPath)
	require.NoError(t, err)

	require.Equal(t, dataRetrieved, data)
}
