package sgx

import (
	"fmt"
	"io/ioutil"

	"github.com/edgelesssys/ego/ecrypto"
	log "github.com/sirupsen/logrus"
)

func SealToFile(data []byte, filePath string) error {
	sealed, err := ecrypto.SealWithProductKey(data, nil)
	if err != nil {
		return fmt.Errorf("failed to seal oracle key: %w", err)
	}

	if err := ioutil.WriteFile(filePath, sealed, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is written successfully", filePath)

	return nil
}
