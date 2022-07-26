package sgx

import (
	"fmt"
	"io/ioutil"

	"github.com/edgelesssys/ego/ecrypto"
	log "github.com/sirupsen/logrus"
)

// Seal returns data sealed with unique ID in SGX-enabled environments
// If SGX is disabled, it returns the data as is.
func Seal(data []byte, enclaveEnabled bool) ([]byte, error) {
	if enclaveEnabled {
		return ecrypto.SealWithUniqueKey(data, nil)
	} else {
		return data, nil
	}
}

func SealToFile(data []byte, filePath string, enclaveEnabled bool) error {
	sealedData, err := Seal(data, enclaveEnabled)
	if err != nil {
		return fmt.Errorf("failed to seal oracle private key: %w", err)
	}

	if err := ioutil.WriteFile(filePath, sealedData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is written successfully", filePath)

	return nil
}
