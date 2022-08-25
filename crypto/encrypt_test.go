package crypto

import (
	"bytes"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
	"testing"
)

// Success encryption and decryption
func TestEncryptData(t *testing.T) {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	pubKey := privKey.PubKey()
	origData := []byte("encrypt origData please")

	cipherText, err := Encrypt(pubKey, origData)
	require.NoError(t, err)

	plainText, err := btcec.Decrypt(privKey, cipherText)

	if !bytes.Equal(origData, plainText) {
		t.Errorf("decrypted data doesn't match original data")
	}
}

// Success encryption but fail decryption
func TestEncryptData_FailDecryption(t *testing.T) {
	privKey1, err1 := btcec.NewPrivateKey(btcec.S256())
	privKey2, err2 := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err1)
	require.NoError(t, err2)

	// encrypt to pubKey1
	pubKey := privKey1.PubKey()
	origData := []byte("decryption will be failed")

	cipherText, err := Encrypt(pubKey, origData)
	require.NoError(t, err)

	// try to decrypt using privKey2
	_, err = btcec.Decrypt(privKey2, cipherText)
	require.Error(t, err)
}
