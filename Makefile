APP ?= nv
BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/$(APP)
MAIN_PKG ?= ./cmd/nv

GOFLAGS ?= -trimpath -buildvcs=false
GOCACHE_DIR ?= $(CURDIR)/.gocache
export GOCACHE := $(GOCACHE_DIR)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS ?= -X github.com/jcouture/nv/internal/cli.Version=$(VERSION) -X github.com/jcouture/nv/internal/cli.Commit=$(COMMIT)
BUILD_FLAGS ?= $(strip $(if $(LDFLAGS),-ldflags "$(LDFLAGS)"))

TEST_PKGS ?= ./...
FUZZ_PKGS ?= ./...
FUZZTIME ?= 30s

.PHONY: build clean help fmt fix vet gosec vulncheck tidy precommit test install uninstall print-version fuzz

.DEFAULT_GOAL := help

## Run unit tests
test:
	@go run gotest.tools/gotestsum@v1.13.0 --format=testdox -- -coverprofile=coverage.out -covermode=atomic $(TEST_PKGS)
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

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

## Build the binary
build:
	@mkdir -p $(BIN_DIR) $(GOCACHE_DIR)
	@CGO_ENABLED=0 go build $(GOFLAGS) $(BUILD_FLAGS) -o $(BIN) $(MAIN_PKG)
	@echo "Built $(BIN)"

## Print the computed release version
print-version:
	@printf '%s\n' "$(VERSION)"

## Install to GOBIN
install:
	@CGO_ENABLED=0 go install $(GOFLAGS) $(BUILD_FLAGS) $(MAIN_PKG)
	@echo "Installed $(APP) to $$(go env GOBIN)"

## Uninstall from GOBIN
uninstall:
	@INSTALL_PATH="$$(go env GOBIN)/$(APP)"; \
	if [ -f "$$INSTALL_PATH" ]; then \
		rm -f "$$INSTALL_PATH"; \
		echo "Uninstalled $(APP) from $$INSTALL_PATH"; \
	else \
		echo "$(APP) not found at $$INSTALL_PATH"; \
	fi

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
	@go get -u ./...
	@go mod tidy -v

## Pre-commit checks (writes fmt/tidy)
precommit: fmt fix tidy vet gosec vulncheck test
	@echo "Pre-commit checks passed"

## Clean build artifacts
clean:
	@echo "GOCACHE_DIR=$(GOCACHE_DIR)"
	@rm -rf "$(BIN_DIR)" dist/ "$(GOCACHE_DIR)"
	@if [ -d .gomodcache ]; then chmod -R u+w .gomodcache && rm -rf .gomodcache; fi
	@go clean -cache -testcache
	@echo "Cleaned build artifacts"

## Show help
help:
	@printf '%s\n\n' "$(APP) - Available targets:"
	@awk 'BEGIN { esc=sprintf("%c", 27); cyan=esc "[36m"; reset=esc "[0m" } /^##/{help=$$0; sub(/^## */, "", help); next} /^[[:alnum:]_.-]+:/{target=$$1; sub(/:.*/, "", target); if(help){printf "  %s%-18s%s %s\n", cyan, target, reset, help; help=""}}' $(MAKEFILE_LIST)
