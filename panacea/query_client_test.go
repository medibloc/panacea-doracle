package panacea_test

// All the tests can only work in sgx environment, so the tests are commented out.
import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// Test for GetAccount function.
//func TestGetAccount(t *testing.T) {
//
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	trustedBlockinfo := panacea.TrustedBlockInfo{
//		TrustedBlockHeight: 99,
//		TrustedBlockHash:   hash,
//	}
//	userHomeDir, err := os.UserHomeDir()
//	require.NoError(t, err)
//
//	homeDir := filepath.Join(userHomeDir, ".doracle")
//	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
//	require.NoError(t, err)
//
//	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
//
//	require.NoError(t, err)
//
//	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
//	accAddr, err := queryClient.GetAccount(mediblocLimitedAddress)
//	require.NoError(t, err)
//
//	address, err := bech32.ConvertAndEncode("panacea", accAddr.GetPubKey().Address().Bytes())
//	require.NoError(t, err)
//
//	require.Equal(t, mediblocLimitedAddress, address)
//}

//func TestMultiGetAddress(t *testing.T) {
//
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	trustedBlockinfo := panacea.TrustedBlockInfo{
//		TrustedBlockHeight: 99,
//		TrustedBlockHash:   hash,
//	}
//	userHomeDir, err := os.UserHomeDir()
//	require.NoError(t, err)
//
//	homeDir := filepath.Join(userHomeDir, ".doracle")
//	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
//	require.NoError(t, err)
//
//	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
//
//	require.NoError(t, err)
//
//	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
//
//	var wg sync.WaitGroup
//	wg.Add(5)
//
//	fmt.Println("query 1 send")
//	go func() {
//		_, err := queryClient.GetAccount(mediblocLimitedAddress)
//		require.NoError(t, err)
//		wg.Done()
//		fmt.Println("query 1 receive")
//	}()
//
//	fmt.Println("query 2 send")
//	go func() {
//		_, err := queryClient.GetAccount(mediblocLimitedAddress)
//		require.NoError(t, err)
//		wg.Done()
//		fmt.Println("query 2 receive")
//	}()
//
//	fmt.Println("query 3 send")
//	go func() {
//		_, err := queryClient.GetAccount(mediblocLimitedAddress)
//		require.NoError(t, err)
//		wg.Done()
//		fmt.Println("query 3 receive")
//	}()
//
//	fmt.Println("query 4 send")
//	go func() {
//		_, err := queryClient.GetAccount(mediblocLimitedAddress)
//		require.NoError(t, err)
//		wg.Done()
//		fmt.Println("query 4 receive")
//	}()
//
//	fmt.Println("query 5 send")
//	go func() {
//		_, err := queryClient.GetAccount(mediblocLimitedAddress)
//		require.NoError(t, err)
//		wg.Done()
//		fmt.Println("query 5 receive")
//	}()
//	wg.Wait()
//
//}

//func TestLightClientConnection(t *testing.T) {
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	userHomeDir, err := os.UserHomeDir()
//	require.NoError(t, err)
//
//	homeDir := filepath.Join(userHomeDir, ".doracle")
//	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
//	require.NoError(t, err)
//
//	trustedBlockinfo := panacea.TrustedBlockInfo{
//		TrustedBlockHeight: 99,
//		TrustedBlockHash:   hash,
//	}
//
//	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
//	require.NoError(t, err)
//
//	_, err = queryClient.LightClient.LastTrustedHeight()
//	require.NoError(t, err)
//
//	_, err = panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
//	require.Error(t, err)
//
//	err = queryClient.Close()
//	require.NoError(t, err)
//
//	queryClient2, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
//	require.NoError(t, err)
//
//	_, err = queryClient2.LightClient.LastTrustedHeight()
//	require.NoError(t, err)
//
//}

func TestGetOracleRegistration(t *testing.T) {
	hash, err := hex.DecodeString("2D83F26DB2E2997659E88ECDE968DCF85EA1F581B5032D770FECC4A44CA5C102")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := panacea.TrustedBlockInfo{
		TrustedBlockHeight: 10,
		TrustedBlockHash:   hash,
	}

	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := filepath.Join(userHomeDir, ".doracle")
	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
	require.NoError(t, err)

	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)

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
