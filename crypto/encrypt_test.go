package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

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

func TestEncryptWithAES256(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	data := []byte("test data")

	shareKey1 := SharedKey(privKey1, privKey2.PubKey())
	shareKey2 := SharedKey(privKey2, privKey1.PubKey())

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)

	encryptedData1, err := EncryptWithAES256(shareKey1, nonce, data)
	require.NoError(t, err)
	encryptedData2, err := EncryptWithAES256(shareKey2, nonce, data)
	require.NoError(t, err)

	require.Equal(t, encryptedData1, encryptedData2)
}

func TestDecryptWithAES256(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	data := []byte("This is temporary data")

	shareKey1 := SharedKey(privKey1, privKey2.PubKey())
	shareKey2 := SharedKey(privKey2, privKey1.PubKey())

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)

	encryptedData, err := EncryptWithAES256(shareKey1, nonce, data)
	require.NoError(t, err)

	decryptedData, err := DecryptWithAES256(shareKey2, nonce, encryptedData)
	require.NoError(t, err)

	require.Equal(t, decryptedData, data)
}
