package panacea_test

import (
	"context"
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// All the tests can only work in sgx environment, so the tests are commented out.

// Test for GetAccount function.
func TestGetAccount(t *testing.T) {

	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := panacea.TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}

	err = panacea.SaveTrustedBlockInfo(trustedBlockinfo)
	require.NoError(t, err)

	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := filepath.Join(userHomeDir, ".doracle")
	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
	require.NoError(t, err)

	gettrustedBlockinfo, err := panacea.GetTrustedBlockInfo()
	require.NoError(t, err)

	queryClient, err := panacea.NewQueryClient(ctx, conf, gettrustedBlockinfo)

	require.NoError(t, err)

	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
	accAddr, err := queryClient.GetAccount(mediblocLimitedAddress)
	require.NoError(t, err)

	address, err := bech32.ConvertAndEncode("panacea", accAddr.GetPubKey().Address().Bytes())
	require.NoError(t, err)

	require.Equal(t, mediblocLimitedAddress, address)

}

//func TestGobBuffer(t *testing.T) {
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	fmt.Println(hash)
//	trustedBlockinfo := panacea.TrustedBlockInfo{
//		TrustedBlockHeight: 99,
//		TrustedBlockHash:   hash,
//	}
//	var buffer bytes.Buffer
//	//var buffer2 bytes.Buffer
//	enc := gob.NewEncoder(&buffer)
//	dec := gob.NewDecoder(&buffer)
//	err = enc.Encode(trustedBlockinfo)
//	require.NoError(t, err)
//
//	//fmt.Println(buffer.Bytes())
//
//	var getBlockinfo panacea.TrustedBlockInfo
//
//	err = dec.Decode(&getBlockinfo)
//	require.NoError(t, err)
//	fmt.Println(getBlockinfo.TrustedBlockHash)
//	fmt.Println(getBlockinfo.TrustedBlockHeight)
//}

// Test for GetBalance function.
// The test fails due to a version problem of the current panacea mainNet.
//func TestGetBalance(t *testing.T) {
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	trustedBlockinfo := panacea.TrustedBlockInfo{
//		TrustedBlockHeight: 99,
//		TrustedBlockHash:   hash,
//	}
//
//	userHomeDir, err := os.UserHomeDir()
//	if err != nil {
//		panic(err)
//	}
//	homeDir := filepath.Join(userHomeDir, ".doracle")
//	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
//	require.NoError(t, err)
//
//	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
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
//	trustedBlockinfo := panacea.TrustedBlockInfo{
//		TrustedBlockHeight: 99,
//		TrustedBlockHash:   hash,
//	}
//
//	userHomeDir, err := os.UserHomeDir()
//	if err != nil {
//		panic(err)
//	}
//	homeDir := filepath.Join(userHomeDir, ".doracle")
//	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
//	require.NoError(t, err)
//
//	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
//
//	require.NoError(t, err)
//
//	mediblocLimitedAddress := "panacea1crvw2ysrlrtzyk0m2u9m0eq0jrmpf6exxx7sex"
//	topic, err := queryClient.GetTopic(mediblocLimitedAddress, "test")
//	require.NoError(t, err)
//
//	fmt.Println("topic: ", topic.String())
//}
