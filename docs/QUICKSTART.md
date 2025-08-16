# Quick Start Guide 🚀

Get up and running with GoChain in minutes! This guide will help you set up your development environment and start building blockchain applications.

## ⚡ **5-Minute Setup**

### **Prerequisites**

- **Go 1.21+** (latest stable recommended)
- **Git**
- **Basic Go knowledge**

### **1. Clone and Setup**

```bash
# Clone the repository
git clone https://github.com/gochain/gochain.git
cd gochain

# Install dependencies
go mod download

# Verify installation
go version
```

### **2. Run Tests (Verify Everything Works)**

```bash
# Run comprehensive test suite
./scripts/test_suite.sh

# Or run individual tests
go test ./pkg/block/... -v
```

### **3. Start Development**

```bash
# Run specific package tests
go test ./pkg/miner/... -v

# Check test coverage
go test -cover ./pkg/block/
```

## 🏗️ **Project Structure Overview**

```
gochain/
├── cmd/                    # Application entry points
├── pkg/                   # Core packages
│   ├── block/            # Block structure & validation [93.0% coverage]
│   ├── chain/            # Blockchain management
│   ├── consensus/        # Consensus mechanisms
│   ├── contracts/        # Smart contract engine
│   ├── defi/             # DeFi protocols
│   ├── miner/            # Mining operations [100% test success]
│   ├── wallet/           # Wallet management
│   └── ...               # More packages
├── scripts/               # Development tools
└── docs/                  # Documentation
```

## 🧪 **Testing Your First Component**

### **Block Package Example**

```go
package main

import (
    "fmt"
    "github.com/gochain/gochain/pkg/block"
)

func main() {
    // Create a new block
    header := &block.Header{
        Version:    1,
        PrevHash:   make([]byte, 32),
        MerkleRoot: make([]byte, 32),
        Timestamp:  time.Now().Unix(),
        Difficulty: 1,
        Nonce:      0,
    }
    
    block := &block.Block{
        Header:       header,
        Transactions: []*block.Transaction{},
    }
    
    // Validate the block
    err := block.Validate()
    if err != nil {
        fmt.Printf("Block validation failed: %v\n", err)
        return
    }
    
    fmt.Println("Block created and validated successfully!")
}
```

## 🔧 **Common Development Tasks**

### **Running Tests**

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/block/... -v

# With coverage
go test -coverprofile=coverage.out ./pkg/block/
go tool cover -html=coverage.out -o coverage.html

# Race detection
go test -race ./...

# Fuzz testing
go test -fuzz=Fuzz ./pkg/wallet/
```

### **Building Components**

```bash
# Build main application
go build ./cmd/gochain

# Build test runner
go build ./cmd/test_runner

# Run benchmarks
go test -bench=. ./pkg/benchmark/
```

### **Code Quality**

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check for race conditions
go test -race ./...
```

## 📚 **Next Steps**

### **For Blockchain Developers**
1. Read [Architecture Guide](ARCHITECTURE.md)
2. Explore [Smart Contract Development](SMART_CONTRACTS.md)
3. Study [Testing Guide](TESTING.md)

### **For DeFi Developers**
1. Read [DeFi Development Guide](DEFI_DEVELOPMENT.md)
2. Explore token standards and AMM protocols
3. Study smart contract security

### **For Researchers**
1. Read [Research Tools](RESEARCH_TOOLS.md)
2. Explore cryptographic implementations
3. Study consensus mechanisms

## 🚨 **Troubleshooting**

### **Common Issues**

**Test Failures**
```bash
# Check Go version
go version

# Clean and rebuild
go clean -cache
go mod download

# Run tests with verbose output
go test -v ./pkg/block/...
```

**Build Errors**
```bash
# Update dependencies
go mod tidy

# Check Go version compatibility
go version

# Clean build artifacts
go clean
```

**Coverage Issues**
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./pkg/block/

# View detailed coverage
go tool cover -func=coverage.out

# Open HTML report
go tool cover -html=coverage.out -o coverage.html
```

## 📞 **Getting Help**

- **Documentation**: Check the [docs/](docs/) folder
- **Issues**: Open an issue on GitHub
- **Discussions**: Use GitHub Discussions
- **Email**: Contact the development team

## 🎯 **Success Metrics**

You're ready to continue when:

- ✅ All tests pass (`./scripts/test_suite.sh`)
- ✅ You can run individual package tests
- ✅ You understand the project structure
- ✅ You can build and run components
- ✅ You have a development environment set up

---

**Ready to build the future of blockchain?** 🚀🔬

**Last Updated**: December 2024  
**Version**: 1.0.0  
**GoChain**: Research-grade blockchain development platform 🚀🔬💻
