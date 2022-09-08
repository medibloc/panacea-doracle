package panacea

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/stretchr/testify/require"
)

// All the tests can only work in sgx environment, so the tests are commented out.

// Test for GetAccount function.
func TestGetAccount(t *testing.T) {

	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}
	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := filepath.Join(userHomeDir, ".doracle")
	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
	require.NoError(t, err)

	queryClient, err := NewQueryClient(ctx, conf, trustedBlockinfo)

	require.NoError(t, err)

	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
	accAddr, err := queryClient.GetAccount(mediblocLimitedAddress)
	require.NoError(t, err)

	address, err := bech32.ConvertAndEncode("panacea", accAddr.GetPubKey().Address().Bytes())
	require.NoError(t, err)

	require.Equal(t, mediblocLimitedAddress, address)

	err = queryClient.Close()
	require.NoError(t, err)
}

func TestMultiGetAddress(t *testing.T) {

	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}
	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := filepath.Join(userHomeDir, ".doracle")
	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
	require.NoError(t, err)

	queryClient, err := NewQueryClient(ctx, conf, trustedBlockinfo)

	require.NoError(t, err)

	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"

	var wg sync.WaitGroup
	wg.Add(5)

	fmt.Println("query 1 send")
	go func() {
		_, err := queryClient.GetAccount(mediblocLimitedAddress)
		require.NoError(t, err)
		wg.Done()
		fmt.Println("query 1 receive")
	}()

	fmt.Println("query 2 send")
	go func() {
		_, err := queryClient.GetAccount(mediblocLimitedAddress)
		require.NoError(t, err)
		wg.Done()
		fmt.Println("query 2 receive")
	}()

	fmt.Println("query 3 send")
	go func() {
		_, err := queryClient.GetAccount(mediblocLimitedAddress)
		require.NoError(t, err)
		wg.Done()
		fmt.Println("query 3 receive")
	}()

	fmt.Println("query 4 send")
	go func() {
		_, err := queryClient.GetAccount(mediblocLimitedAddress)
		require.NoError(t, err)
		wg.Done()
		fmt.Println("query 4 receive")
	}()

	fmt.Println("query 5 send")
	go func() {
		_, err := queryClient.GetAccount(mediblocLimitedAddress)
		require.NoError(t, err)
		wg.Done()
		fmt.Println("query 5 receive")
	}()
	wg.Wait()

	err = queryClient.Close()
	require.NoError(t, err)

}

func TestLoadQueryClient(t *testing.T) {
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := filepath.Join(userHomeDir, ".doracle")
	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
	require.NoError(t, err)

	trustedBlockinfo := TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}

	queryClient, err := NewQueryClient(ctx, conf, trustedBlockinfo)
	require.NoError(t, err)

	_, err = queryClient.lightClient.LastTrustedHeight()
	require.NoError(t, err)

	_, err = NewQueryClient(ctx, conf, trustedBlockinfo)
	require.Error(t, err)

	err = queryClient.Close()
	require.NoError(t, err)

	queryClient, err = LoadQueryClient(ctx, conf)
	require.NoError(t, err)

	_, err = queryClient.lightClient.LastTrustedHeight()
	require.NoError(t, err)

	err = queryClient.Close()
	require.NoError(t, err)

}
