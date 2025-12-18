APP ?= nv

BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/$(APP)

GOFLAGS ?= -trimpath -buildvcs=false
GOEXPERIMENT ?= greenteagc

GOCACHE_DIR ?= $(CURDIR)/.gocache
export GOCACHE := $(GOCACHE_DIR)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS ?= -w -s -X main.Version=$(VERSION) -X main.GitCommit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)
BUILD_FLAGS ?= $(strip $(if $(LDFLAGS),-ldflags "$(LDFLAGS)"))

TEST_PKGS ?= ./...

.PHONY: build clean help fmt vet tidy precommit test install release-dry-run run go-mod-update tag

.DEFAULT_GOAL := help

## Run unit tests
test:
	@GOEXPERIMENT=$(GOEXPERIMENT) go test -v -cover $(TEST_PKGS)

## Build the nv binary
build:
	@mkdir -p $(BIN_DIR) $(GOCACHE_DIR)
	@go mod download && go mod verify
	@GOEXPERIMENT=$(GOEXPERIMENT) CGO_ENABLED=0 go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN) ./cmd/nv
	@echo "Built $(BIN)"

## Install nv to GOPATH/bin
install:
	@GOEXPERIMENT=$(GOEXPERIMENT) CGO_ENABLED=0 go install $(GOFLAGS) $(BUILD_FLAGS) ./cmd/nv
	@echo "Installed $(APP) to $$(go env GOPATH)/bin"

## Format code (writes changes)
fmt:
	@go fmt ./...
	@find . -name '*.go' -not -path './.*' | xargs -r gofmt -s -w
	@echo "Code formatted"

## Static analysis (vet)
vet:
	@go vet ./...
	@echo "Vet passed"

## Tidy modules (writes go.mod/go.sum if needed)
tidy:
	@go mod tidy -v

## Local pre-commit convenience (writes fmt/tidy)
precommit: fmt tidy vet test
	@echo "Pre-commit checks passed"

## Update dependencies
go-mod-update:
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated"

## Dry-run GoReleaser
release-dry-run:
	@goreleaser --snapshot --skip=publish --clean

## Launch the executable with optional arguments (use: make run ARGS="...")
run:
	@GOEXPERIMENT=$(GOEXPERIMENT) go run ./cmd/nv/nv.go $(ARGS)

## Git tag a version (use: make tag VERSION=v1.0.0)
tag:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag VERSION=v1.0.0"; exit 1; fi
	@git tag -a $(VERSION) -s -m "$(VERSION)"
	@echo "Tagged $(VERSION)"

## Clean build artifacts
clean:
	@rm -rf $(BIN_DIR) $(GOCACHE_DIR)
	@go clean
	@echo "Cleaned build artifacts"

## Show this help message
help:
	@echo "$(APP) - Available targets:"
	@echo ""
	@awk '/^##/{help=$$0; sub(/^## */, "", help); next} /^[[:alnum:]_.-]+:/{target=$$1; sub(/:.*/, "", target); if(help){printf "  \033[36m%-18s\033[0m %s\n", target, help; help=""}}' $(MAKEFILE_LIST)
