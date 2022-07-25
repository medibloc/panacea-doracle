export GO111MODULE = on

GO ?= ego-go

build_tags := $(strip $(BUILD_TAGS))
BUILD_FLAGS := -tags "$(build_tags)"

OUT_DIR = ./build

.PHONY: all build test install ego-sign clean

all: build test install

build: go.sum
	$(GO) build -mod=readonly $(BUILD_FLAGS) -o $(OUT_DIR)/doracled ./cmd/doracled

ego-sign:
	ego sign enclave.json

test:
	$(GO) test -v ./...

install: go.sum
	$(GO) install -mod=readonly $(BUILD_FLAGS) ./cmd/doracled

clean:
	$(GO) clean
	rm -rf $(OUT_DIR)
