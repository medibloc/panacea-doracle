package panacea

import (
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type TxBuilder struct {
	client QueryClient
}

func NewTxBuilder(client QueryClient) *TxBuilder {
	return &TxBuilder{
		client: client,
	}
}

// GenerateSignedTxBytes signs msgs using the private key and returns the signed Tx message in form of byte array.
func (tb TxBuilder) GenerateSignedTxBytes(
	privateKey cryptotypes.PrivKey,
	gasLimit uint64,
	feeAmount sdk.Coins,
	msg ...sdk.Msg,
) ([]byte, error) {
	txConfig := authtx.NewTxConfig(tb.client.cdc, []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})
	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(feeAmount)

	if err := txBuilder.SetMsgs(msg...); err != nil {
		return nil, err
	}

	signerAddress, err := bech32.ConvertAndEncode(HRP, privateKey.PubKey().Address().Bytes())
	if err != nil {
		return nil, err
	}

	signerAccount, err := tb.client.GetAccount(signerAddress)
	if err != nil {
		return nil, err
	}

	sigV2 := signing.SignatureV2{
		PubKey: privateKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: signerAccount.GetSequence(),
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       tb.client.chainID,
		AccountNumber: signerAccount.GetAccountNumber(),
		Sequence:      signerAccount.GetSequence(),
	}

	sigV2, err = clienttx.SignWithPrivKey(
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

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	return txConfig.TxEncoder()(txBuilder.GetTx())
}
