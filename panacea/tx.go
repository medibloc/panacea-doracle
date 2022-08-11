package panacea

import (
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type TxBuilder struct {
	client     GrpcClientI
	marshaller *codec.ProtoCodec
	chainID    string
}

func NewTxBuilder(client GrpcClientI) *TxBuilder {
	marshaller := codec.NewProtoCodec(client.GetInterfaceRegistry())

	return &TxBuilder{
		client:     client,
		marshaller: marshaller,
	}
}

func (tb TxBuilder) GenerateSignedTxBytes(
	privateKey cryptotypes.PrivKey,
	gasLimit uint64,
	msg ...sdk.Msg,
) ([]byte, error) {
	txConfig := authtx.NewTxConfig(tb.marshaller, []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})
	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetGasLimit(gasLimit)

	if err := txBuilder.SetMsgs(msg...); err != nil {
		return nil, err
	}

	signerAddress, err := bech32.ConvertAndEncode(Hrp, privateKey.PubKey().Address().Bytes())
	if err != nil {
		return nil, err
	}

	signerAccount, err := tb.client.GetAccount(signerAddress)
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       tb.client.GetChainID(),
		AccountNumber: signerAccount.GetAccountNumber(),
		Sequence:      signerAccount.GetSequence(),
	}

	sigV2, err := clienttx.SignWithPrivKey(
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder,
		privateKey,
		txConfig,
		signerAccount.GetSequence(),
	)
	if err != nil {
		return nil, err
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	return txConfig.TxEncoder()(txBuilder.GetTx())
}
