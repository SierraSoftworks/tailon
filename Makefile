# Makefile for tailon

.PHONY: build build-all test test-verbose test-coverage clean run help

# Default target
all: build

# Build the application
build:
	go build -o tailon .

# Cross-compile for all supported platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=windows GOARCH=amd64 go build -o tailon-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o tailon-windows-arm64.exe .
	GOOS=linux GOARCH=amd64 go build -o tailon-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o tailon-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o tailon-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o tailon-darwin-arm64 .
	@echo "Cross-compilation complete!"
	@echo "Built binaries:"
	@ls -la tailon-*

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific package tests
test-config:
	go test -v ./pkg/config

test-apps:
	go test -v ./pkg/apps

test-api:
	go test -v ./pkg/api

test-main:
	go test -v .

# Run tests without Tailscale (using build tags)
test-no-tailscale:
	go test -tags=no_tailscale -v ./...

# Clean build artifacts
clean:
	rm -f tailon
	rm -f tailon-*
	rm -f coverage.out
	rm -f coverage.html

# Run the application with default config
run:
	go run main.go

# Run with verbose logging
run-verbose:
	go run main.go --verbose

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  build-all       - Cross-compile for all supported platforms"
	@echo "  test            - Run all tests"
	@echo "  test-verbose    - Run tests with verbose output"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-config     - Run config package tests"
	@echo "  test-apps       - Run apps package tests"
	@echo "  test-api        - Run API package tests"
	@echo "  test-main       - Run main package tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  run             - Run the application"
	@echo "  run-verbose     - Run with verbose logging"
	@echo "  deps            - Install dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  help            - Show this help"
