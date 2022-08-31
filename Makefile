export GO111MODULE = on

GO ?= ego-go

build_tags := $(strip $(BUILD_TAGS))
BUILD_FLAGS := -tags "$(build_tags)"

OUT_DIR = ./build

.PHONY: all build test clean

all: build test

build: go.sum
	$(GO) build -mod=readonly $(BUILD_FLAGS) -o $(OUT_DIR)/doracled ./cmd/doracled

test:
ifeq ($(GO),ego-go)
	./scripts/run-tests-with-ego.sh
else
	$(GO) test -v ./...
endif

sign-prod: build
	ego sign ./scripts/enclave-prod.json

clean:
	$(GO) clean
	rm -rf $(OUT_DIR)
