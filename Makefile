GO       ?= go
BIN_DIR  ?= bin
PKG      ?= ./...

ifeq ($(OS),Windows_NT)
BINARY := orvix.exe
else
BINARY := orvix
endif

OUTPUT := $(BIN_DIR)/$(BINARY)

.PHONY: all build test vet tidy clean run help

all: build

build: | $(BIN_DIR)
	$(GO) build -o $(OUTPUT) .

test:
	$(GO) test $(PKG)

vet:
	$(GO) vet $(PKG)

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

run: build
	./$(OUTPUT) $(ARGS)

help:
	@echo "Targets:"
	@echo "  build  Compile orvix into $(BIN_DIR)/"
	@echo "  test   Run unit tests"
	@echo "  vet    Run go vet"
	@echo "  tidy   Sync go.mod / go.sum"
	@echo "  clean  Remove $(BIN_DIR)/"
	@echo "  run    Build, then run with ARGS, e.g. make run ARGS=\"submit --dry-run examples/hello.sh\""

$(BIN_DIR):
	mkdir -p $(BIN_DIR)
