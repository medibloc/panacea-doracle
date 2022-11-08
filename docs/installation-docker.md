# Installation: Use Docker

NOTE: This installation process works only on the hardwares specified in the [Hardware Requirements](../README.md#hardware-requirements)


## Pull an image

```bash
docker pull ghcr.io/medibloc/panacea-doracle:latest
```
It is highly recommended to use a specific Docker image tag instead of `latest`. You can find image tags from the [Github Packages page](https://github.com/medibloc/panacea-doracle/pkgs/container/panacea-doracle).


## Run a container using SGX

This is a sample command to show you how to run a container using SGX in your host.

```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v ${ANY_DIR_ON_HOST}:/doracle \
    ghcr.io/medibloc/panacea-doracle:latest \
    ego run /usr/bin/doracled --help
```