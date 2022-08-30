package crypto

import "github.com/btcsuite/btcd/btcec"

// Encrypt encrypts data using a secp256k1 public key (ECIES)
func Encrypt(pubKey *btcec.PublicKey, data []byte) ([]byte, error) {
	return btcec.Encrypt(pubKey, data)
}

// Decrypt decrypts data using a secp256k1 private key (ECIES)
func Decrypt(privKey *btcec.PrivateKey, data []byte) ([]byte, error) {
	return btcec.Decrypt(privKey, data)
}
