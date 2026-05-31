# ==============================================================================
# BIG-IP Exporter — Makefile
# ==============================================================================

BINARY      := f5_bigip_exporter
PKG         := ./cmd/f5_bigip_exporter
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(DATE)

GO          ?= go
GOFLAGS     ?=
DOCKER_IMG  ?= f5_bigip_exporter

.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/## //' | awk -F': ' '{printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## build: Build the binary
.PHONY: build
build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

## run: Run the exporter locally (uses config-example.yml)
.PHONY: run
run: build
	./$(BINARY) -config config-example.yml -insecure

## test: Run unit tests
.PHONY: test
test:
	$(GO) test -race -count=1 ./...

## cover: Run tests with coverage report
.PHONY: cover
cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

## fmt: Format the code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

## vet: Run go vet
.PHONY: vet
vet:
	$(GO) vet ./...

## lint: Run golangci-lint (must be installed)
.PHONY: lint
lint:
	golangci-lint run ./...

## tidy: Tidy go modules
.PHONY: tidy
tidy:
	$(GO) mod tidy

## check: Run fmt, vet, and test (CI gate)
.PHONY: check
check: fmt vet test

## docker: Build the Docker image
.PHONY: docker
docker:
	docker build -t $(DOCKER_IMG):$(VERSION) -t $(DOCKER_IMG):latest .

## clean: Remove build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY) coverage.out coverage.html
