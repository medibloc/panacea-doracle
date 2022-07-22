package secp256k1

import "github.com/btcsuite/btcd/btcec"

func NewPrivKey() (*btcec.PrivateKey, error) {
	return btcec.NewPrivateKey(btcec.S256())
}
