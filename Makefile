export GO111MODULE = on

GO ?= ego-go

build_tags := $(strip $(BUILD_TAGS))
BUILD_FLAGS := -tags "$(build_tags)"

OUT_DIR = ./build

.PHONY: all build test sign-prod clean

all: build test

build: go.sum
	$(GO) build -mod=readonly $(BUILD_FLAGS) -o $(OUT_DIR)/doracled ./cmd/doracled

test:
	./scripts/run-tests-with-ego.sh

# Prepare ./scripts/private.pem that you want to use. If not, this command will generate a new one.
sign-prod:
	ego sign ./scripts/enclave-prod.json

clean:
	$(GO) clean
	rm -rf $(OUT_DIR)
