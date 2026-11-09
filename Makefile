GO           ?= go
PYTHON       ?= python3
BIN_DIR      ?= bin
PKG          ?= ./...
# CGO disabled by default: produce a fully static binary with no libc
# dependency, so it runs on systems with older glibc (e.g. the HPC).
CGO_ENABLED  ?= 0

BINARY := orvix

OUTPUT := $(BIN_DIR)/$(BINARY)

GO_BUILD_FLAGS ?=

.PHONY: all build build-all build-linux-amd64 build-linux-arm64 test regen-golden vet tidy clean run help docs-directives docs-check docs-html

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

# Regenerate jobspec conformance goldens from the current build.
# Use after an intentional change to generated output, and review the
# resulting diff under spec/jobspec/conformance/golden/ in the same commit.
regen-golden:
	$(GO) test ./spec/jobspec/ -run TestConformanceGolden -update

vet:
	$(GO) vet $(GO_BUILD_FLAGS) $(PKG)

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

run: build
	./$(OUTPUT) $(ARGS)

# Regenerate the directive reference tables from the jobspec schema
# (spec/jobspec/jobspec.schema.json) plus docs/data/directives.json.
# Commit the regenerated fragment together with the contract change.
docs-directives:
	$(PYTHON) docs/tools/generate_directive_tables.py

# Drift gate used by CI: the committed fragment must match the schema and data.
docs-check:
	$(PYTHON) docs/tools/generate_directive_tables.py --check

# Strict documentation build (warnings as errors), includes the drift gate.
docs-html: docs-check
	$(PYTHON) -m sphinx -b html -W --keep-going -a -E docs/source docs/build/html

help:
	@echo "Targets:"
	@echo "  build                  Compile orvix for the current platform"
	@echo "  build-all              Compile orvix for the current platform and all Linux targets"
	@echo "  build-linux-amd64      Cross-compile for Linux AMD64"
	@echo "  build-linux-arm64      Cross-compile for Linux ARM64"
	@echo "  test                   Run unit tests"
	@echo "  regen-golden           Regenerate jobspec conformance goldens (review the diff)"
	@echo "  vet                    Run go vet"
	@echo "  tidy                   Sync go.mod / go.sum"
	@echo "  clean                  Remove $(BIN_DIR)/"
	@echo "  run                    Build, then run with ARGS, e.g. make run ARGS=\"submit --dry-run examples/hello.sh\""
	@echo "  release-snapshot       GoReleaser dry-run (release artifacts to dist/, no publish)"
	@echo "  docs-directives        Regenerate the directive reference tables from the schema"
	@echo "  docs-check             Check the committed directive tables for drift"
	@echo "  docs-html              Strict docs build (warnings as errors), includes docs-check"
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
