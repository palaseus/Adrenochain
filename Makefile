# GoChain Blockchain - Makefile
# Provides convenient commands for building, testing, and managing the project

.PHONY: help build test test-verbose test-coverage test-race test-fuzz test-bench test-all clean install deps lint format check

# Default target
help:
	@echo "🚀 GoChain Blockchain - Available Commands:"
	@echo ""
	@echo "📦 Building:"
	@echo "  build          - Build the GoChain binary"
	@echo "  install        - Install GoChain binary to GOPATH"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test           - Run all tests (fast)"
	@echo "  test-verbose   - Run all tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage reporting"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-fuzz      - Run fuzz tests only"
	@echo "  test-bench     - Run benchmark tests only"
	@echo "  test-all       - Run comprehensive test suite (recommended)"
	@echo ""
	@echo "🔧 Development:"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  lint           - Run linter checks"
	@echo "  format         - Format Go code"
	@echo "  check          - Run all checks (lint + format + test)"
	@echo ""
	@echo "📊 Analysis:"
	@echo "  coverage       - Generate coverage report"
	@echo "  security       - Run security checks"
	@echo "  performance    - Run performance benchmarks"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make test-all  # Run everything and get detailed report"

# Build the project
build:
	@echo "🔨 Building GoChain..."
	go build -o bin/gochain ./cmd/gochain
	@echo "✅ Build complete: bin/gochain"

# Install to GOPATH
install:
	@echo "📦 Installing GoChain..."
	go install ./cmd/gochain
	@echo "✅ Installation complete"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf test_results/
	rm -rf coverage/
	go clean -cache -testcache
	@echo "✅ Clean complete"

# Download and tidy dependencies
deps:
	@echo "📥 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies updated"

# Run basic tests
test:
	@echo "🧪 Running tests..."
	go test ./... -timeout 30s

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	go test ./... -v -timeout 30s

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -covermode=atomic -timeout 30s
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "🧪 Running tests with race detection..."
	go test ./... -race -timeout 60s

# Run fuzz tests
test-fuzz:
	@echo "🧪 Running fuzz tests..."
	@find ./pkg -name "*_fuzz_test.go" -exec dirname {} \; | sort -u | while read pkg; do \
		echo "Testing $$pkg..."; \
		go test -fuzz=Fuzz -fuzztime=30s "$$pkg"; \
	done

# Run benchmark tests
test-bench:
	@echo "📊 Running benchmark tests..."
	@find ./pkg -name "*_test.go" -exec grep -l "Benchmark" {} \; | xargs dirname | sort -u | while read pkg; do \
		echo "Benchmarking $$pkg..."; \
		go test -bench=. -benchmem "$$pkg"; \
	done

# Run comprehensive test suite (recommended)
test-all:
	@echo "🚀 Running comprehensive test suite..."
	@if [ -f "scripts/test_suite.sh" ]; then \
		./scripts/test_suite.sh; \
	else \
		echo "❌ Test suite script not found. Running basic tests..."; \
		make test-verbose; \
	fi

# Generate coverage report
coverage:
	@echo "📊 Generating coverage report..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Run linter checks
lint:
	@echo "🔍 Running linter checks..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	elif command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "⚠️  No linter found. Install golangci-lint or golint."; \
		exit 1; \
	fi

# Format Go code
format:
	@echo "🎨 Formatting Go code..."
	go fmt ./...
	@echo "✅ Code formatting complete"

# Run all checks
check: format lint test
	@echo "✅ All checks passed!"

# Security checks
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

# Performance benchmarks
performance:
	@echo "📊 Running performance benchmarks..."
	make test-bench
	@echo "✅ Performance analysis complete"

# Development setup
setup:
	@echo "🚀 Setting up GoChain development environment..."
	make deps
	make format
	make test
	@echo "✅ Development environment ready!"

# Quick validation
validate:
	@echo "✅ Validating GoChain..."
	go build ./...
	go test ./... -timeout 10s
	@echo "✅ Validation complete"

# Show project status
status:
	@echo "📊 GoChain Project Status:"
	@echo "  📦 Go version: $(shell go version)"
	@echo "  📁 Project root: $(shell pwd)"
	@echo "  🔧 Go modules: $(shell go list -m)"
	@echo "  🧪 Test packages: $(shell go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./... | wc -l)"
	@echo "  📊 Coverage: $(shell if [ -f coverage.out ]; then go tool cover -func=coverage.out | tail -1; else echo "Not available"; fi)"