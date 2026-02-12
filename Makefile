BINARY_NAME=salja
PKG=github.com/gongahkia/salja
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

.PHONY: build test lint fmt vet coverage install clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/salja

test:
	go test ./... -v

lint:
	@which golangci-lint > /dev/null 2>&1 || echo "golangci-lint not installed, skipping"
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || true

fmt:
	gofmt -w .

vet:
	go vet ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

install:
	go install $(LDFLAGS) ./cmd/salja

clean:
	rm -rf bin/ coverage.out coverage.html
