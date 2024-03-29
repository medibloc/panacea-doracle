package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
)

// Encrypt encrypts data using a secp256k1 public key (ECIES)
func Encrypt(pubKey *btcec.PublicKey, data []byte) ([]byte, error) {
	return btcec.Encrypt(pubKey, data)
}

// Decrypt decrypts data using a secp256k1 private key (ECIES)
func Decrypt(privKey *btcec.PrivateKey, data []byte) ([]byte, error) {
	return btcec.Decrypt(privKey, data)
}

// EncryptWithAES256 encrypts data using a AES256 cryptography.
func EncryptWithAES256(secretKey, nonce, data []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aesgcm.NonceSize() {
		return nil, fmt.Errorf("nonce length must be %v", aesgcm.NonceSize())
	}

	cipherText := aesgcm.Seal(nil, nonce, data, nil)

	return cipherText, nil
}

// DecryptWithAES256 decrypts data using a AES256 cryptography.
func DecryptWithAES256(secretKey, nonce, ciphertext []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
