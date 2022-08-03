package panacea_test

import (
	"context"
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
	"testing"
)

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
