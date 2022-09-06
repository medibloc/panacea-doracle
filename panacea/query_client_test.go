package panacea_test

// All the tests can only work in sgx environment, so the tests are commented out.
import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
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
	hash, err := hex.DecodeString("93EF7A66EE58C1063B13CE408D0D850CE2B4E396D366C92BD5DB1BBF9FA1C4BC")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := panacea.TrustedBlockInfo{
		TrustedBlockHeight: 10,
		TrustedBlockHash:   hash,
	}

	conf := &config.Config{
		Panacea: config.PanaceaConfig{
			ChainID:       "local",
			RpcAddr:       "tcp://127.0.0.1:26657",
			PrimaryAddr:   "tcp://127.0.0.1:26657",
			WitnessesAddr: "tcp://127.0.0.1:26657",
			GRPCAddr:      "127.0.0.1:9090",
		},
	}

	grpcClient, err := panacea.NewGrpcClient(conf)
	require.NoError(t, err)
	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
	require.NoError(t, err)

	mnemonic := "genre cook grace border huge learn collect suffer head casino trial elegant hood check organ galaxy athlete become super typical bulk describe scout fetch"
	oracleAccount, err := panacea.NewOracleAccount(mnemonic, 0, 0)
	require.NoError(t, err)

	//// get unique ID
	//selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	//require.NoError(t, err)
	//uniqueID := hex.EncodeToString(selfEnclaveInfo.UniqueID)
	//fmt.Println("uniqueID: ", uniqueID)

	oracleRegistrationFromGrpc, err := grpcClient.GetOracleRegistration(oracleAccount.GetAddress(), "41dbf6cf1f732b23765c0ad3d2282225e7f02ce185ba639fb1f1e746ca4ae677")
	require.NoError(t, err)

	fmt.Println(oracleRegistrationFromGrpc)

	oracleRegistration, err := queryClient.GetOracleRegistration(oracleAccount.GetAddress(), "41dbf6cf1f732b23765c0ad3d2282225e7f02ce185ba639fb1f1e746ca4ae677")
	require.NoError(t, err)

	fmt.Println(oracleRegistration)
}
