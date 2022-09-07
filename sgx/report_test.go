package sgx_test

import (
	"testing"

	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyRemoteReport(t *testing.T) {
	data := []byte("hello")

	report, err := sgx.GenerateRemoteReport(data)
	require.NoError(t, err)
	require.NotEmpty(t, report)

	enclaveInfo, err := sgx.GetSelfEnclaveInfo()
	require.NoError(t, err)
	require.NotNil(t, enclaveInfo)
	require.NotEmpty(t, enclaveInfo.ProductID)
	require.NotEmpty(t, enclaveInfo.SignerID)
	require.NotEmpty(t, enclaveInfo.UniqueID)

	err = sgx.VerifyRemoteReport(report, data, *enclaveInfo)
	require.NoError(t, err)
}

func TestVerifyRemoteReportWithDifferentData(t *testing.T) {
	data := []byte("data")

	report, err := sgx.GenerateRemoteReport(data)
	require.NoError(t, err)
	require.NotEmpty(t, report)

	enclaveInfo, err := sgx.GetSelfEnclaveInfo()
	require.NoError(t, err)
	require.NotNil(t, enclaveInfo)

	wrongData := []byte("wrong data")

	err = sgx.VerifyRemoteReport(report, wrongData, *enclaveInfo)
	require.Error(t, err)
}
