package sgx

import (
	"github.com/edgelesssys/ego/enclave"
	log "github.com/sirupsen/logrus"
)

type EnclaveInfo struct {
	ProductID []byte
	SignerID  []byte
	UniqueID  []byte
}

func NewEnclaveInfo(productID, signerID, uniqueID []byte) *EnclaveInfo {
	return &EnclaveInfo{
		ProductID: productID,
		SignerID:  signerID,
		UniqueID:  uniqueID,
	}
}

// GetSelfEnclaveInfo sets EnclaveInfo from self-generated remote report
func GetSelfEnclaveInfo() (*EnclaveInfo, error) {
	// generate self-remote-report and get product ID, signer ID, and unique ID
	reportBz, err := GenerateRemoteReport([]byte(""))
	if err != nil {
		log.Errorf("failed to generate self-report: %v", err)
		return nil, err
	}

	report, err := enclave.VerifyRemoteReport(reportBz)
	if err != nil {
		log.Errorf("failed to retrieve self-report: %v", err)
		return nil, err
	}

	return NewEnclaveInfo(report.ProductID, report.SignerID, report.UniqueID), nil
}
