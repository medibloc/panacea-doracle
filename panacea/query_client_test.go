package panacea_test

import (
	"context"
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
	"testing"
)

// All the tests can only work in sgx environment, so the tests are commented out.

// Test for GetAccount function.
func TestGetAccount(t *testing.T) {
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	queryClient, err := panacea.NewQueryClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)

	require.NoError(t, err)

	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
	accAddr, err := queryClient.GetAccount(mediblocLimitedAddress)
	require.NoError(t, err)

	address, err := bech32.ConvertAndEncode("panacea", accAddr.GetPubKey().Address().Bytes())
	require.NoError(t, err)

	require.Equal(t, mediblocLimitedAddress, address)

}

// Test for GetBalance function.
// The test fails due to a version problem of the current panacea mainNet.
//func TestGetBalance(t *testing.T) {
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	queryClient, err := panacea.NewQueryClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)
//
//	require.NoError(t, err)
//
//	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
//	balance, err := queryClient.GetBalance(mediblocLimitedAddress)
//	require.NoError(t, err)
//
//	fmt.Println("balance: ", balance.String())
//
//}

// Test for GetTopic function.
// It is commented out because it is a test in a local environment.
//func TestGetTopicLocal(t *testing.T) {
//	hash, err := hex.DecodeString("226F43C4D9962545285E736B64004A83528E36281DB8CC4B7A1C60FECA003832")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	queryClient, err := panacea.NewQueryClient(ctx, "local", "http://127.0.0.1:26657", 99, hash)
//
//	require.NoError(t, err)
//
//	mediblocLimitedAddress := "panacea1crvw2ysrlrtzyk0m2u9m0eq0jrmpf6exxx7sex"
//	topic, err := queryClient.GetTopic(mediblocLimitedAddress, "test")
//	require.NoError(t, err)
//
//	fmt.Println("topic: ", topic.String())
//}
