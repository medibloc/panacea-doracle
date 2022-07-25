# Panacea doracle

A decentralized oracle which validates off-chain data to be transacted in the data exchange protocol of the Panacea chain while preserving privacy

## Features

- Validating that data meets the requirements of a specific deal
    - with utilizing TEE (Trusted Execution Environment) for preserving privacy
- Providing encrypted data to buyers


## Hardware Requirements

The oracle only works on [SGX](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)-[FLC](https://github.com/intel/linux-sgx/blob/master/psw/ae/ref_le/ref_le.md) environment with a [quote provider](https://docs.edgeless.systems/ego/#/reference/attest) installed.
You can check if your hardware supports SGX and it is enabled in the BIOS by following [EGo guide](https://docs.edgeless.systems/ego/#/getting-started/troubleshoot?id=hardware).


## Prerequisites

### Install prerequisites

```bash
sudo apt update
sudo apt install build-essential libssl-dev

sudo snap install go --classic
sudo snap install ego-dev --classic
sudo ego install az-dcap-client

sudo usermod -a -G sgx_prv $USER
```

### Setting signing key

To make ego-sign, you have to prepare a signing key.  

```bash
openssl genrsa -out private.pem -3 3072
openssl rsa -in private.pem -pubout -out public.pem
```

### Setting `enclave.json`

An example of `enclave.json` is given.
A source directory would be mounted as `/data` in SGX.

```json
{
  "exe": "./build/doracled",
  "key": "private.pem",
  "debug": true,
  "heapSize": 512,
  "executableHeap": false,
  "productID": 1,
  "securityVersion": 1,
  "mounts": [
    {
      "source": "<a-directory-you-want>",
      "target": "/data",
      "type": "hostfs",
      "readOnly": false
    },
    {
      "target": "/tmp",
      "type": "memfs"
    }
  ],
  "env": [
    {
      "name": "HOME",
      "value": "/data"
    }
  ],
  "files": null
}
```

## Build

```bash
# in SGX-enabled environment,
make build

# in SGX-disabled environment,
GO=go make build
```

## EGo Sign

EGo sign is executable only in SGX-enabled environments.

```bash
make ego-sign
```

## Test

```bash
# in SGX-enabled environment,
make test

# in SGX-disabled environment,
GO=go make test
```

## Installation

```bash
# in SGX-enabled environment,
make install

# in SGX-disabled environment,
GO=go make install
```

## Initialize

```bash
doracled init
```

## Run

Before running the binary, the environment variable `SGX_AESM_ADDR` must be unset.
If not, the Azure DCAP client won't be used automatically.
```bash
unset SGX_AESM_ADDR
# If not,
#
# ERROR: sgxquoteexprovider: failed to load libsgx_quote_ex.so.1: libsgx_quote_ex.so.1: cannot open shared object file: No such file or directory [openenclave-src/host/sgx/linux/sgxquoteexloader.c:oe_sgx_load_quote_ex_library:118]
# ERROR: Failed to load SGX quote-ex library (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/host/sgx/sgxquote.c:oe_sgx_qe_get_target_info:688]
# ERROR: SGX Plugin _get_report(): failed to get ecdsa report. OE_QUOTE_LIBRARY_LOAD_ERROR (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/enclave/sgx/attester.c:_get_report:320]
```

Run the binary using `ego` so that it can be run in the secure enclave.
### generate oracle key
```bash
# For the first oracle that generates an oracle key,
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracled gen-oracle-key
```

Then, two files are generated under `~/.doracle/`
- `oracle_priv_key.sealed` : sealed oracle private key
- `oracle_pub_key.json` : oracle public key & its remote report
