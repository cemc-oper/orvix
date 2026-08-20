GO           ?= go
BIN_DIR      ?= bin
PKG          ?= ./...
# CGO disabled by default: produce a fully static binary with no libc
# dependency, so it runs on systems with older glibc (e.g. the HPC).
CGO_ENABLED  ?= 0

BINARY := orvix

OUTPUT := $(BIN_DIR)/$(BINARY)

GO_BUILD_FLAGS ?=

.PHONY: all build build-all build-linux-amd64 build-linux-arm64 test vet tidy clean run help

all: build

build: | $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GO_BUILD_FLAGS) -o $(OUTPUT) .

# Linux cross-compilation targets (CGO disabled for fully static binaries,
# so they run on HPC systems with older glibc)
build-linux-amd64: | $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-linux-amd64 .

build-linux-arm64: | $(BIN_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/orvix-linux-arm64 .

build-all: build build-linux-amd64 build-linux-arm64

test:
	$(GO) test $(GO_BUILD_FLAGS) $(PKG)

vet:
	$(GO) vet $(GO_BUILD_FLAGS) $(PKG)

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

run: build
	./$(OUTPUT) $(ARGS)

help:
	@echo "Targets:"
	@echo "  build                  Compile orvix for the current platform"
	@echo "  build-all              Compile orvix for the current platform and all Linux targets"
	@echo "  build-linux-amd64      Cross-compile for Linux AMD64"
	@echo "  build-linux-arm64      Cross-compile for Linux ARM64"
	@echo "  test                   Run unit tests"
	@echo "  vet                    Run go vet"
	@echo "  tidy                   Sync go.mod / go.sum"
	@echo "  clean                  Remove $(BIN_DIR)/"
	@echo "  run                    Build, then run with ARGS, e.g. make run ARGS=\"submit --dry-run examples/hello.sh\""
	@echo "  release-snapshot       GoReleaser dry-run (release artifacts to dist/, no publish)"
	@echo ""
	@echo "Variables:"
	@echo "  GO_BUILD_FLAGS         Extra flags for go build/test/vet"
	@echo "  CGO_ENABLED            0 (default) = fully static binary; 1 = link against system libc"

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# GoReleaser dry-run: build release artifacts locally without publishing
# (requires goreleaser on PATH; output goes to dist/)
.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean
