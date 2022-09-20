package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// CreateTLSConfig creates a x509 certificate and generate a rsa key.
func CreateTLSConfig() (*tls.Config, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "PanaceaOracle"},
		NotAfter:     time.Now().AddDate(1, 0, 0),
		DNSNames:     []string{"localhost"}, // TODO: Set proper DNS names
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			{Certificate: [][]byte{certBytes}, PrivateKey: priv},
		},
	}, nil
}
