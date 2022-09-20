#!/bin/bash

# If a test pkg requires SGX,
#   - compiles a test binary for each Go package that has `*_test.go` files.
#     - because Go doesn't allow us to build a single test binary for all packages.
#   - signs the test binary with EGo.
#   - runs the test binary with EGo.
# If not, runs the test pkg in the usual way.

set -euo pipefail

unset SGX_AESM_ADDR

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT=${SCRIPT_DIR}/..
TEST_BIN=${ROOT}/test.bin
GARBAGES="${TEST_BIN} enclave.json private.pem public.pem"

SGX_TEST_PKGS=(
  "github.com/medibloc/panacea-doracle/sgx"
)

arr_contains() {
  local array="$1[@]"
  local target=$2
  for element in "${!array}"; do
    if [[ $element == "$target" ]]; then
      return 0
    fi
  done
  return 1
}

TEST_PKGS=$(go list -test ${ROOT}/... | grep '\.test$' | sed 's|\.test$||g')
for PKG in ${TEST_PKGS}; do
  if arr_contains SGX_TEST_PKGS ${PKG} ; then  # if SGX is required
    if [ "${GO}" != "ego-go" ]; then  # if SGX isn't enabled
      continue
    fi

    rm -f ${GARBAGES}
    ${GO} test -mod=readonly -c -o ${TEST_BIN} ${PKG}  # builds a test binary (without running tests)
    ego sign ${TEST_BIN}  # generates private/public.pem and enclave.json automatically
    ego run ${TEST_BIN} -test.v
  else
    ${GO} test -v -count=1 ${PKG}
  fi
done

rm -f ${GARBAGES}
