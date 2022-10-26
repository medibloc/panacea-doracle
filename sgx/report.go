package sgx

import (
	"bytes"
	"fmt"
	"github.com/edgelesssys/ego/enclave"
)

// GenerateRemoteReport generates a SGX report containing the specified data for use in remote attestation.
// This works only in the SGX-FLC environment where the SGX quote provider is installed.
func GenerateRemoteReport(data []byte) ([]byte, error) {
	return enclave.GetRemoteReport(data)
}

// VerifyRemoteReport verifies whether the report not only was properly generated in the SGX environment,
// but also contains the promised security version, product ID, unique ID and signer ID,
// in order to verify that the report was generated by the promised binary which was not forged.
func VerifyRemoteReport(reportBytes, expectedData []byte, expectedEnclaveInfo EnclaveInfo) error {
	report, err := enclave.VerifyRemoteReport(reportBytes)
	if err != nil {
		return err
	}

	if report.SecurityVersion < PromisedMinSecurityVersion {
		return fmt.Errorf("invalid security version in the report")
	}
	if !bytes.Equal(report.ProductID, expectedEnclaveInfo.ProductID) {
		return fmt.Errorf("invalid product ID in the report")
	}
	if !bytes.Equal(report.SignerID, expectedEnclaveInfo.SignerID) {
		return fmt.Errorf("invalid signer ID in the report")
	}
	if !bytes.Equal(report.UniqueID, expectedEnclaveInfo.UniqueID) {
		return fmt.Errorf("invalid unique ID in the report")
	}
	if !bytes.Equal(report.Data[:len(expectedData)], expectedData) {
		return fmt.Errorf("invalid data in the report")
	}

	return nil
}
