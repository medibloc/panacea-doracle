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

```bash
sudo apt update
sudo apt install build-essential libssl-dev

sudo snap install go --classic
sudo snap install ego-dev --classic
sudo ego install az-dcap-client

sudo usermod -a -G sgx_prv $USER
# After this, reopen the shell so that the updated Linux group info can be loaded.
```


## Build a `doracled` binary

```bash
# in SGX-enabled environment,
make build

# in SGX-disabled environment,
GO=go make build
```


## Run unit tests

```bash
# in SGX-enabled environment,
make test

# in SGX-disabled environment,
GO=go make test
```


## Sign the `doracled` with Ego

To run the binary in the enclave, the binary must be signed with EGo.

### In development

First of all, prepare a RSA private key of the signer.

```bash
openssl genrsa -out private.pem -3 3072
openssl rsa -in private.pem -pubout -out public.pem
```

After that, prepare a `enclave.json` file that will be applied when signing the binary with EGo.

In the following example, please replace the `<a-directory-you-want>` with a directory in your local.
This will be mounted as `/home_mnt` to the file system presented to the enclave.
And, the `HOME` environment variable will indicate the `/home_mnt`.

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
      "target": "/home_mnt",
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
      "value": "/home_mnt"
    }
  ],
  "files": null
}
```

Finally, you can sign the binary with EGo.

```bash
ego sign ./enclave.json
```

If the binary is signed successfully, you can move the binary to where you want.

### In production

A configuration for production is already prepared in the [`scripts/enclave-prod.json`](scripts/enclave-prod.json).

So, you can just put your RSA private key (`private.pem`) into the `scripts/`, and run the following command to sign the binary.

```bash
ego sign ./scripts/enclave-prod.json
```

If the binary is signed successfully, you can move the binary to where you want, or publish the binary to GitHub or so.

Note that a `/doracle` directory must be created and its permissions must be set properly before running the binary,
because the `/doracle` directory will be mounted as a `HOME` directory to the enclave.
For more details, please see the [`scripts/enclave-prod.json`](scripts/enclave-prod.json).


## Initialize directories for the `doracled`

```bash
doracled init
```

## Run the `doracled`

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

### Generate an oracle key

```bash
# For the first oracle that generates an oracle key,
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracled gen-oracle-key
```

Then, two files are generated under `~/.doracle/`
- `oracle_priv_key.sealed` : sealed oracle private key
- `oracle_pub_key.json` : oracle public key & its remote report


### Verify the remote report

You can verify that key is generated in SGX using the promised binary.
For that, the public key and its remote report are required.

```json
{
  "public_key_base_64" : "<base64-encoded-public-key>",
  "remote_report_base_64": "<base64-encoded-remote-report>"
}
```

Then, you can verify the remote report.

```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracled verify-report <remote-report-path>
```

### Register an oracle to the Panacea

Request to register an oracle.

The trusted block information is required which would be used for light client verification.
The account number and index are optional with the default value of 0.

```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracled register-oracle \
--trusted-block-height <block-height> \
--trusted-block-hash <base64-encoded-block-hash> \
--acc-num <account-number> \
--index <account-index>
```

### Get the oracle key registered in the Panacea

If an oracle registered successfully (vote for oracle registration is passed), the oracle can be shared the oracle private key.
The oracle private key is encrypted and shared, and it can only be decrypted using the node private key (which is used when registering oracle) 

```bash
AZDCAP_DEBUG_LOG_LEVEL=INFO ego run doracled get-oracle-key
```

The oracle private key is sealed and stored in a file named `oracle_priv_key.sealed` under `~/.doracle/`
