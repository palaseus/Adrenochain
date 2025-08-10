.PHONY: help build test clean lint format coverage bench fuzz run

# Default target
help:
	@echo "GoChain - Available commands:"
	@echo "  build     - Build the GoChain binary"
	@echo "  test      - Run all tests"
	@echo "  test-race - Run tests with race detection"
	@echo "  clean     - Clean build artifacts"
	@echo "  lint      - Run linter"
	@echo "  format    - Format code with gofmt"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  bench     - Run benchmarks"
	@echo "  fuzz      - Run fuzz tests"
	@echo "  run       - Build and run GoChain"
	@echo "  install   - Install GoChain to GOPATH"
	@echo "  proto     - Generate protobuf code"

# Build the binary
build:
	@echo "Building GoChain..."
	go build -o gochain ./cmd/gochain
	@echo "Build complete: ./gochain"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests complete"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race -v ./...
	@echo "Race detection tests complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f gochain
	rm -rf test_data_*
	go clean -cache -testcache
	@echo "Clean complete"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting complete"

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatting complete"

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...
	@echo "Benchmarks complete"

# Run fuzz tests
fuzz:
	@echo "Running fuzz tests..."
	go test -fuzz=. -fuzztime=30s ./...
	@echo "Fuzz tests complete"

# Build and run
run: build
	@echo "Running GoChain..."
	./gochain

# Install to GOPATH
install:
	@echo "Installing GoChain..."
	go install ./cmd/gochain
	@echo "Installation complete"

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/net/message.proto
	@echo "Protobuf generation complete"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@echo "Development setup complete"

# Security check
security:
	@echo "Running security checks..."
	go list -json -deps . | nancy sleuth
	@echo "Security check complete"