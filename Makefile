BINARY_NAME=salja
MCP_BINARY_NAME=salja-mcp
PKG=github.com/gongahkia/salja
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"
LDFLAGS_MCP=-ldflags "-X main.version=$(VERSION)"

.PHONY: build build-mcp test lint fmt vet coverage install clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/salja

build-mcp:
	go build $(LDFLAGS_MCP) -o bin/$(MCP_BINARY_NAME) ./cmd/salja-mcp

test:
	go test ./... -v -timeout 60s

lint:
	@if which golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

fmt:
	gofmt -w .

vet:
	go vet ./...

coverage:
	go test ./... -coverprofile=coverage.out -timeout 60s
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

install:
	go install $(LDFLAGS) ./cmd/salja
	go install $(LDFLAGS_MCP) ./cmd/salja-mcp

clean:
	rm -rf bin/ coverage.out coverage.html
