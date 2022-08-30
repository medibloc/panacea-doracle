package sgx

import (
	"fmt"
	"io/ioutil"

	"github.com/edgelesssys/ego/ecrypto"
	log "github.com/sirupsen/logrus"
)

// SealToFile seals the data with unique ID and stores it to file.
func SealToFile(data []byte, filePath string) error {
	sealedData, err := ecrypto.SealWithUniqueKey(data, nil)
	if err != nil {
		return fmt.Errorf("failed to seal oracle private key: %w", err)
	}

	if err := ioutil.WriteFile(filePath, sealedData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is sealed and written successfully", filePath)

	return nil
}

func UnsealFromFile(filePath string) ([]byte, error) {
	sealed, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	key, err := ecrypto.Unseal(sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle key: %w", err)
	}

	return key, nil
}
