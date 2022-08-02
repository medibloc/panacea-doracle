package secp256k1

import "github.com/btcsuite/btcd/btcec"

// NewPrivKey generates a random secp256k1 private key
func NewPrivKey() (*btcec.PrivateKey, error) {
	return btcec.NewPrivateKey(btcec.S256())
}
