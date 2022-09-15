package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// NewPrivKey generates a random secp256k1 private key
func NewPrivKey() (*btcec.PrivateKey, error) {
	return btcec.NewPrivateKey(btcec.S256())
}

func PrivKeyFromBytes(privKeyBz []byte) (*btcec.PrivateKey, *btcec.PublicKey) {
	return btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)
}

func GeneratePrivateKeyFromMnemonic(mnemonic string, coinType, accNum, index uint32) (secp256k1.PrivKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	hdPath := hd.NewFundraiserParams(accNum, coinType, index).String()
	master, ch := hd.ComputeMastersFromSeed(bip39.NewSeed(mnemonic, ""))

	return hd.DerivePrivateKeyForPath(master, ch, hdPath)
}

func SharedKey(priv *btcec.PrivateKey, pub *btcec.PublicKey) []byte {
	return btcec.GenerateSharedSecret(priv, pub)
}
