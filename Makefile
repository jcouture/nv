BIN_DIR ?= bin
BIN_NV ?= $(BIN_DIR)/nv

GOFLAGS ?= -trimpath -buildvcs=false

GOCACHE_DIR ?= $(CURDIR)/.gocache
export GOCACHE := $(GOCACHE_DIR)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS ?= -X github.com/jcouture/nv/internal/cli.Version=$(VERSION) -X github.com/jcouture/nv/internal/cli.Commit=$(COMMIT) -X github.com/jcouture/nv/internal/build.Version=$(VERSION) -X github.com/jcouture/nv/internal/build.GitCommit=$(COMMIT)
BUILD_FLAGS ?= $(strip $(if $(LDFLAGS),-ldflags "$(LDFLAGS)"))

TEST_PKGS ?= ./...
FUZZ_PKGS ?= ./...
FUZZTIME ?= 30s

.PHONY: build clean help fmt fix vet gosec vulncheck tidy precommit test install fuzz coverage

.DEFAULT_GOAL := help

## Run unit tests
test:
	@go run gotest.tools/gotestsum@v1.13.0 --format=testdox -- -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

.PHONY: coverage
coverage: test
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run fuzz tests
fuzz:
	@for pkg in $$(go list $(FUZZ_PKGS)); do \
		fuzzes=$$(go test $$pkg -list '^Fuzz' 2>/dev/null | grep '^Fuzz'); \
		if [ -z "$$fuzzes" ]; then \
			continue; \
		fi; \
		for fuzz in $$fuzzes; do \
			echo "Fuzzing $$pkg ($$fuzz)"; \
			go test $$pkg -run=^$$ -fuzz=$$fuzz -fuzztime=$(FUZZTIME); \
		done; \
	done

## Build nv binary
build:
	@mkdir -p $(BIN_DIR) $(GOCACHE_DIR)
	CGO_ENABLED=0 go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN_NV) ./cmd/nv
	@echo "Built $(BIN_NV)"

## Install nv to GOPATH/bin
install:
	CGO_ENABLED=0 go install $(GOFLAGS) $(BUILD_FLAGS) ./cmd/nv
	@echo "Installed nv to $$(go env GOPATH)/bin"

## Format code (writes changes)
fmt:
	@go fmt ./...
	@find . -name '*.go' -not -path './.*' | xargs -r gofmt -s -w
	@echo "Code formatted"

## Apply automated Go fixes
fix:
	@go fix ./...
	@echo "Go fix applied"

## Static analysis (vet)
vet:
	@go vet ./...
	@echo "Vet passed"

## Security analysis (gosec)
gosec:
	@go run github.com/securego/gosec/v2/cmd/gosec@v2.22.1 ./...
	@echo "Gosec passed"

## Vulnerability scanning
vulncheck:
	@go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
	@echo "Vulnerability scan passed"

## Tidy modules (writes go.mod/go.sum if needed)
tidy:
	@go mod tidy -v

## Local pre-commit convenience (writes fmt/tidy)
precommit: fmt fix tidy vet gosec vulncheck test
	@echo "Pre-commit checks passed"

## Clean build artifacts
clean:
	@echo "GOCACHE_DIR=$(GOCACHE_DIR)"
	@rm -rf "$(BIN_DIR)" dist/ "$(GOCACHE_DIR)"
	@go clean -cache -testcache
	@echo "Cleaned build artifacts"

## Show this help message
help:
	@echo "nv - Available targets:"
	@echo ""
	@awk '/^##/{help=$$0; sub(/^## */, "", help); next} /^[[:alnum:]_.-]+:/{target=$$1; sub(/:.*/, "", target); if(help){printf "  \033[36m%-18s\033[0m %s\n", target, help; help=""}}' $(MAKEFILE_LIST)
