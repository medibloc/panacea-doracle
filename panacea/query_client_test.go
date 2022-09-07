package panacea_test

// All the tests can only work in sgx environment, so the tests are commented out.
import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Test for GetAccount function.
func TestGetAccount(t *testing.T) {

	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := panacea.TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}

	conf := &config.Config{
		BaseConfig: config.BaseConfig{
			LogLevel:          "",
			OracleMnemonic:    "",
			ListenAddr:        "",
			Subscriber:        "",
			DataDir:           "data",
			OraclePrivKeyFile: "oracle_priv_key.sealed",
			OraclePubKeyFile:  "oracle_pub_key.json",
			NodePrivKeyFile:   "node_priv_key.sealed",
		},
		Panacea: config.PanaceaConfig{
			GRPCAddr:                "https://grpc.gopanacea.org:443",
			RPCAddr:                 "https://rpc.gopanacea.org:443",
			ChainID:                 "panacea-3",
			DefaultGasLimit:         200000,
			DefaultFeeAmount:        "1000000umed",
			LightClientPrimaryAddr:  "https://rpc.gopanacea.org:443",
			LightClientWitnessAddrs: []string{"https://rpc.gopanacea.org:443"},
		},
	}

	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)

	require.NoError(t, err)

	mediblocLimitedAddress := "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3"
	accAddr, err := queryClient.GetAccount(mediblocLimitedAddress)
	require.NoError(t, err)

	address, err := bech32.ConvertAndEncode("panacea", accAddr.GetPubKey().Address().Bytes())
	require.NoError(t, err)

	require.Equal(t, mediblocLimitedAddress, address)
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

	trustedBlockinfo := panacea.TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}

	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
	require.NoError(t, err)

	_, err = queryClient.LightClient.LastTrustedHeight()
	require.NoError(t, err)

	_, err = panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
	require.Error(t, err)

	err = queryClient.Close()
	require.NoError(t, err)

	queryClient, err = panacea.LoadQueryClient(ctx, conf)
	require.NoError(t, err)

	_, err = queryClient.LightClient.LastTrustedHeight()
	require.NoError(t, err)

	err = queryClient.Close()
	require.NoError(t, err)

}

func TestMultiGetAddress(t *testing.T) {

	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	require.NoError(t, err)
	ctx := context.Background()

	trustedBlockinfo := panacea.TrustedBlockInfo{
		TrustedBlockHeight: 99,
		TrustedBlockHash:   hash,
	}
	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := filepath.Join(userHomeDir, ".doracle")
	conf, err := config.ReadConfigTOML(filepath.Join(homeDir, "config.toml"))
	require.NoError(t, err)

	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)

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

}

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
			ChainID:                 "local",
			RPCAddr:                 "tcp://127.0.0.1:26657",
			LightClientPrimaryAddr:  "tcp://127.0.0.1:26657",
			LightClientWitnessAddrs: []string{"tcp://127.0.0.1:26657"},
			GRPCAddr:                "127.0.0.1:9090",
		},
	}

	grpcClient, err := panacea.NewGrpcClient(conf)
	require.NoError(t, err)
	queryClient, err := panacea.NewQueryClient(ctx, conf, trustedBlockinfo)
	require.NoError(t, err)

	mnemonic := "genre cook grace border huge learn collect suffer head casino trial elegant hood check organ galaxy athlete become super typical bulk describe scout fetch"

	oracleAccount, err := panacea.NewOracleAccount(mnemonic, 0, 0)
	require.NoError(t, err)

	// generate node key and its remote report
	nodePubKey, nodePubKeyRemoteReport, err := generateNodeKey()
	require.NoError(t, err)

	report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
	uniqueID := hex.EncodeToString(report.UniqueID)

	// sign and broadcast to Panacea
	msgRegisterOracle := oracletypes.NewMsgRegisterOracle(uniqueID, oracleAccount.GetAddress(), nodePubKey, nodePubKeyRemoteReport, trustedBlockinfo.TrustedBlockHeight, trustedBlockinfo.TrustedBlockHash)

	txBuilder := panacea.NewTxBuilder(grpcClient)

	defaultFeeAmount, _ := sdk.ParseCoinsNormalized("1500000umed")
	txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccount.GetPrivKey(), 300000, defaultFeeAmount, msgRegisterOracle)
	require.NoError(t, err)

	_, err = grpcClient.BroadcastTx(txBytes)
	require.NoError(t, err)

	fmt.Println("register-oracle transaction succeed")

	oracleRegistrationFromGrpc, err := grpcClient.GetOracleRegistration(oracleAccount.GetAddress(), uniqueID)
	require.NoError(t, err)

	fmt.Println("unique ID1:", oracleRegistrationFromGrpc.UniqueId)

	oracleRegistration, err := queryClient.GetOracleRegistration(oracleAccount.GetAddress(), uniqueID)
	require.NoError(t, err)

	fmt.Println("unique ID2:", oracleRegistration.UniqueId)

	require.EqualValues(t, oracleRegistrationFromGrpc.UniqueId, oracleRegistration.UniqueId)
}

func generateNodeKey() ([]byte, []byte, error) {
	nodePrivKey, err := crypto.NewPrivKey()
	if err != nil {
		return nil, nil, err
	}
	nodePubKey := nodePrivKey.PubKey().SerializeCompressed()
	oraclePubKeyHash := sha256.Sum256(nodePubKey)
	nodeKeyRemoteReport, err := sgx.GenerateRemoteReport(oraclePubKeyHash[:])
	if err != nil {
		return nil, nil, err
	}

	return nodePubKey, nodeKeyRemoteReport, nil
}
