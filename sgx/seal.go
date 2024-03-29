package sgx

import (
	"fmt"
	"os"

	"github.com/edgelesssys/ego/ecrypto"
	log "github.com/sirupsen/logrus"
)

// SealToFile seals the data with unique ID and stores it to file.
func SealToFile(data []byte, filePath string) error {
	sealedData, err := ecrypto.SealWithUniqueKey(data, nil)
	if err != nil {
		return fmt.Errorf("failed to seal oracle private key: %w", err)
	}

	if err := os.WriteFile(filePath, sealedData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is sealed and written successfully", filePath)

	return nil
}

func UnsealFromFile(filePath string) ([]byte, error) {
	sealed, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	key, err := ecrypto.Unseal(sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle key: %w", err)
	}

	return key, nil
}

// Seal returns data sealed with unique ID in SGX-enabled environments
// If SGX is disabled, it returns the data as is.
func Seal(data []byte, enclaveEnabled bool) ([]byte, error) {
	if enclaveEnabled {
		return ecrypto.SealWithUniqueKey(data, nil)
	} else {
		return data, nil
	}
}

// Unseal returns data unsealed with unique ID in SGX-enabled environments
// If SGX is disabled, it returns the data as is.
func Unseal(data []byte, enclaveEnabled bool) ([]byte, error) {
	if enclaveEnabled {
		return ecrypto.Unseal(data, nil)
	} else {
		return data, nil
	}
}
