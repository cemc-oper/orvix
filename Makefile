GO       ?= go
BIN_DIR  ?= bin
PKG      ?= ./...

ifeq ($(OS),Windows_NT)
BINARY := orvix.exe
else
BINARY := orvix
endif

OUTPUT := $(BIN_DIR)/$(BINARY)

# Auto-enable vendor mode if vendor/modules.txt exists (offline HPC builds)
GO_BUILD_FLAGS ?=
ifneq (,$(wildcard vendor/modules.txt))
GO_BUILD_FLAGS += -mod=vendor
endif

.PHONY: all build build-all build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 test vet tidy vendor vendor-clean clean run help

all: build

build: | $(BIN_DIR)
	$(GO) build $(GO_BUILD_FLAGS) -o $(OUTPUT) .

# Cross-compilation targets (CGO disabled for fully static binaries)
build-linux-amd64: | $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-linux-amd64 .

build-linux-arm64: | $(BIN_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-linux-arm64 .

build-darwin-amd64: | $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-darwin-amd64 .

build-darwin-arm64: | $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-darwin-arm64 .

build-windows-amd64: | $(BIN_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-windows-amd64.exe .

build-all: build build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

test:
	$(GO) test $(GO_BUILD_FLAGS) $(PKG)

vet:
	$(GO) vet $(GO_BUILD_FLAGS) $(PKG)

tidy:
	$(GO) mod tidy

vendor:
	$(GO) mod vendor

vendor-clean:
	rm -rf vendor

clean:
	rm -rf $(BIN_DIR)

run: build
	./$(OUTPUT) $(ARGS)

help:
	@echo "Targets:"
	@echo "  build                  Compile orvix for the current platform"
	@echo "  build-all              Compile orvix for all supported platforms"
	@echo "  build-linux-amd64      Cross-compile for Linux AMD64"
	@echo "  build-linux-arm64      Cross-compile for Linux ARM64"
	@echo "  build-darwin-amd64     Cross-compile for macOS AMD64"
	@echo "  build-darwin-arm64     Cross-compile for macOS ARM64 (Apple Silicon)"
	@echo "  build-windows-amd64    Cross-compile for Windows AMD64"
	@echo "  test                   Run unit tests"
	@echo "  vet                    Run go vet"
	@echo "  tidy                   Sync go.mod / go.sum"
	@echo "  vendor                 Create/update vendor/ directory for offline builds"
	@echo "  vendor-clean           Remove vendor/ directory"
	@echo "  clean                  Remove $(BIN_DIR)/"
	@echo "  run                    Build, then run with ARGS, e.g. make run ARGS=\"submit --dry-run examples/hello.sh\""
	@echo ""
	@echo "Variables:"
	@echo "  GO_BUILD_FLAGS         Extra flags for go build/test/vet (auto-set to -mod=vendor when vendor/ exists)"

$(BIN_DIR):
	mkdir -p $(BIN_DIR)
