package sgx

import "github.com/edgelesssys/ego/enclave"

// GenerateRemoteReport generates a SGX report containing the specified data for use in remote attestation.
// This works only in the SGX-FLC environment where the SGX quote provider is installed.
func GenerateRemoteReport(data []byte, enclaveEnabled bool) ([]byte, error) {
	if enclaveEnabled {
		return enclave.GetRemoteReport(data)
	} else {
		return nil, nil
	}
}
