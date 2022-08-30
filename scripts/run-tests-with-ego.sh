#!/bin/bash

# This scripts:
#   - compiles a test binary for each Go package that has `*_test.go` files.
#     - because Go doesn't allow us to build a single test binary for all packages.
#   - signs the test binary with EGo.
#   - runs the test binary with EGo.

set -euxo pipefail

unset SGX_AESM_ADDR

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT=${SCRIPT_DIR}/..
TEST_BIN=${ROOT}/test.bin
GARBAGES="${TEST_BIN} enclave.json private.pem public.pem"

PKGS_WITH_TESTS=$(go list -test ${ROOT}/... | grep '\.test$' | sed 's|\.test$||g')
for PKG in ${PKGS_WITH_TESTS}; do
	# Skip some packages that need to be refactored for CI
	# TODO: refactor these packages
	if [ ${PKG} == "github.com/medibloc/panacea-doracle/panacea" ]; then
		continue
	fi
	
	rm -f ${GARBAGES}

	ego-go test -mod=readonly -c -o ${TEST_BIN} ${PKG}  # build a test binary (without running tests)
	ego sign ${TEST_BIN}  # generates private/public.pem and enclave.json automatically
	ego run ${TEST_BIN} -test.v
done

rm -f ${GARBAGES}
