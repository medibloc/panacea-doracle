package oracle

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/integration/rest"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func Test(t *testing.T) {
	hash, height, err := rest.QueryLatestBlock("http://localhost:1317")
	require.NoError(t, err)

	trustedBlockInfo := &panacea.TrustedBlockInfo{
		TrustedBlockHeight: height,
		TrustedBlockHash:   hash,
	}

	conf := &config.Config{
		Panacea: config.PanaceaConfig{
			DefaultFeeAmount:        "1000000umed",
			DefaultGasLimit:         200000,
			GRPCAddr:                "tcp://localhost:9090",
			RPCAddr:                 "tcp://localhost:26657",
			ChainID:                 "gyuguen-1",
			LightClientPrimaryAddr:  "tcp://localhost:26657",
			LightClientWitnessAddrs: []string{"tcp://localhost:26657"},
		},
	}

	mnemonic := "cereal segment happy comic remember symptom large maple region pig remember vault artefact business marine page reopen hockey vital dentist boost online name ticket"
	mnemonic2 := "dutch rigid sniff clerk ozone marble heart furnace slender bubble attitude fee lizard until betray safe blame evidence axis feature ozone side behind protect"

	privKey1, err := crypto.GeneratePrivateKeyFromMnemonic(mnemonic, panacea.CoinType, 0, 0)
	require.NoError(t, err)
	privKey2, err := crypto.GeneratePrivateKeyFromMnemonic(mnemonic2, panacea.CoinType, 0, 0)
	require.NoError(t, err)

	addr1, err := bech32.ConvertAndEncode(panacea.HRP, privKey1.PubKey().Address().Bytes())
	require.NoError(t, err)
	addr2, err := bech32.ConvertAndEncode(panacea.HRP, privKey2.PubKey().Address().Bytes())
	require.NoError(t, err)

	queryClient, err := panacea.NewQueryClientWithDB(context.Background(), conf, trustedBlockInfo, dbm.NewMemDB())
	require.NoError(t, err)
	grpcClient, err := panacea.NewGrpcClient(conf.Panacea.GRPCAddr)
	require.NoError(t, err)

	txBuilder := panacea.NewTxBuilder(*queryClient)

	msg := &types.MsgSend{
		FromAddress: addr2,
		ToAddress:   addr1,
		Amount:      sdk.NewCoins(sdk.NewCoin("umed", sdk.NewInt(1000000))),
	}

	msg2 := &types.MsgSend{
		FromAddress: addr2,
		ToAddress:   addr1,
		Amount:      sdk.NewCoins(sdk.NewCoin("umed", sdk.NewInt(10000000))),
	}

	msg3 := &types.MsgSend{
		FromAddress: addr2,
		ToAddress:   addr1,
		Amount:      sdk.NewCoins(sdk.NewCoin("umed", sdk.NewInt(10000000))),
	}

	pk := &secp256k1.PrivKey{
		Key: privKey2,
	}

	bytes, err := txBuilder.GenerateTxBytes(pk, conf, msg, msg2, msg3)
	require.NoError(t, err)

	resp, err := grpcClient.BroadcastTx(bytes)
	fmt.Println(resp)
}
