package panacea_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test for GetAccount function.
//func TestGetAccount(t *testing.T) {
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	queryClient, err := panacea.NewQueryClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)
//
//	require.NoError(t, err)
//
//	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
//	accAddr, err := queryClient.GetAccount(mediblocLimitedAddress)
//
//	require.NoError(t, err)
//
//	address, err := bech32.ConvertAndEncode("panacea", accAddr.GetPubKey().Address().Bytes())
//	require.NoError(t, err)
//
//	require.Equal(t, mediblocLimitedAddress, address)
//}

func TestGetOracleRegistration(t *testing.T) {
	hash, err := hex.DecodeString("2D83F26DB2E2997659E88ECDE968DCF85EA1F581B5032D770FECC4A44CA5C102")
	require.NoError(t, err)
	ctx := context.Background()

	queryClient, err := panacea.NewQueryClient(ctx, "local", "127.0.0.1", 10, hash)

	require.NoError(t, err)

	mediblocLimitedAddress := "panacea16xl6zlglk5c4u2qjuf4ds8lp59swv666hp0x5a"
	// get unique ID
	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	require.NoError(t, err)
	uniqueID := base64.StdEncoding.EncodeToString(selfEnclaveInfo.UniqueID)

	oracleRegistration, err := queryClient.GetOracleRegistration(uniqueID, mediblocLimitedAddress)
	require.NoError(t, err)

	fmt.Println(oracleRegistration)
}
