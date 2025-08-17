# GoChain Developer Guide

## Overview

Welcome to the GoChain Developer Guide! This comprehensive guide will help you understand, set up, and develop on the GoChain platform. GoChain is a high-performance blockchain platform designed for DeFi applications, cross-chain interoperability, and enterprise solutions.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Architecture Overview](#architecture-overview)
3. [Development Environment Setup](#development-environment-setup)
4. [Core Concepts](#core-concepts)
5. [Smart Contract Development](#smart-contract-development)
6. [DeFi Protocol Integration](#defi-protocol-integration)
7. [Cross-Chain Bridge Development](#cross-chain-bridge-development)
8. [Governance & DAO Development](#governance--dao-development)
9. [Testing & Deployment](#testing--deployment)
10. [Performance Optimization](#performance-optimization)
11. [Security Best Practices](#security-best-practices)
12. [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites

- **Go 1.21+**: [golang.org/dl](https://golang.org/dl/)
- **Git**: [git-scm.com](https://git-scm.com/)
- **Docker**: [docker.com](https://docker.com/) (optional)
- **Node.js 18+**: [nodejs.org](https://nodejs.org/) (for frontend development)

### Quick Start

1. **Clone the Repository**
   ```bash
   git clone https://github.com/gochain/gochain.git
   cd gochain
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Build the Project**
   ```bash
   go build ./...
   ```

4. **Run Tests**
   ```bash
   go test ./...
   ```

5. **Start Local Node**
   ```bash
   go run cmd/node/main.go
   ```

## Architecture Overview

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │    │   Blockchain    │
│   Applications  │◄──►│   & Load        │◄──►│   Core          │
│                 │    │   Balancer      │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Microservices │
                       │                 │
                       │ • Exchange      │
                       │ • Bridge        │
                       │ • Governance    │
                       │ • DeFi          │
                       └─────────────────┘
```

### Core Components

1. **Blockchain Core**
   - Consensus mechanism (PoS + PoA hybrid)
   - Block validation and propagation
   - Transaction processing
   - State management

2. **Exchange Layer**
   - Order book management
   - Matching engine
   - Trading pairs
   - Market data

3. **Bridge Infrastructure**
   - Cross-chain communication
   - Validator management
   - Asset mapping
   - Security controls

4. **Governance System**
   - Proposal management
   - Voting mechanisms
   - Treasury management
   - DAO operations

5. **DeFi Protocols**
   - Lending pools
   - AMM (Automated Market Maker)
   - Yield farming
   - Liquidity management

## Development Environment Setup

### Local Development

1. **Environment Variables**
   ```bash
   export GOCHAIN_ENV=development
   export GOCHAIN_PORT=8080
   export GOCHAIN_DB_PATH=./data
   export GOCHAIN_LOG_LEVEL=debug
   ```

2. **Database Setup**
   ```bash
   # LevelDB (default)
   mkdir -p data/chainstate
   
   # PostgreSQL (optional)
   docker run -d --name gochain-postgres \
     -e POSTGRES_PASSWORD=gochain \
     -e POSTGRES_DB=gochain \
     -p 5432:5432 postgres:15
   ```

3. **Configuration Files**
   ```yaml
   # config/development.yaml
   server:
     port: 8080
     host: "0.0.0.0"
   
   blockchain:
     network_id: 1337
     genesis_file: "config/genesis.json"
   
   exchange:
     enabled: true
     max_orders: 1000000
   
   bridge:
     enabled: true
     min_validators: 3
   ```

### IDE Setup

#### VS Code
```json
// .vscode/settings.json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"]
}
```

#### GoLand
- Enable Go modules
- Configure Go version
- Set up code inspection rules

### Development Tools

1. **Code Quality**
   ```bash
   # Install tools
   go install golang.org/x/lint/golint@latest
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Run checks
   golangci-lint run
   goimports -w .
   ```

2. **Testing Tools**
   ```bash
   # Run tests with coverage
   go test -cover ./...
   
   # Run benchmarks
   go test -bench=. ./...
   
   # Run race detector
   go test -race ./...
   ```

## Core Concepts

### Blockchain Fundamentals

#### Blocks
```go
type Block struct {
    Header       *BlockHeader
    Transactions []*Transaction
    Hash         common.Hash
    Number       uint64
}
```

#### Transactions
```go
type Transaction struct {
    Nonce    uint64
    GasPrice *big.Int
    GasLimit uint64
    To       *common.Address
    Value    *big.Int
    Data     []byte
    V, R, S  *big.Int
}
```

#### State Management
```go
type StateDB interface {
    GetBalance(addr common.Address) *big.Int
    SetBalance(addr common.Address, balance *big.Int)
    GetCode(addr common.Address) []byte
    SetCode(addr common.Address, code []byte)
    GetState(addr common.Address, hash common.Hash) common.Hash
    SetState(addr common.Address, key, value common.Hash)
}
```

### Consensus Mechanism

GoChain uses a hybrid consensus mechanism combining Proof of Stake (PoS) and Proof of Authority (PoA):

```go
type ConsensusEngine interface {
    Start() error
    Stop() error
    ValidateBlock(block *Block) error
    FinalizeBlock(block *Block) error
}
```

## Smart Contract Development

### Solidity Contracts

#### Basic Token Contract
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract GoChainToken is ERC20 {
    constructor() ERC20("GoChain Token", "GOCH") {
        _mint(msg.sender, 1000000 * 10**decimals());
    }
}
```

#### DeFi Protocol Contract
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract LendingPool is Ownable {
    mapping(address => uint256) public deposits;
    mapping(address => uint256) public borrows;
    
    function deposit(address token, uint256 amount) external {
        IERC20(token).transferFrom(msg.sender, address(this), amount);
        deposits[msg.sender] += amount;
    }
    
    function borrow(address token, uint256 amount) external {
        require(deposits[msg.sender] >= amount, "Insufficient collateral");
        IERC20(token).transfer(msg.sender, amount);
        borrows[msg.sender] += amount;
    }
}
```

### Contract Deployment

1. **Compile Contracts**
   ```bash
   # Using Hardhat
   npx hardhat compile
   
   # Using Truffle
   truffle compile
   ```

2. **Deploy to Local Network**
   ```bash
   # Start local node
   go run cmd/node/main.go
   
   # Deploy contract
   npx hardhat run scripts/deploy.js --network localhost
   ```

3. **Verify Contract**
   ```bash
   npx hardhat verify --network mainnet <contract_address>
   ```

## DeFi Protocol Integration

### Exchange Integration

#### Order Management
```go
// Create order
order := &orderbook.Order{
    ID:          "order_123",
    TradingPair: "BTC/USDT",
    Side:        orderbook.OrderSideBuy,
    Type:        orderbook.OrderTypeLimit,
    Quantity:    big.NewInt(1000000),
    Price:       big.NewInt(50000),
    UserID:      "user123",
}

err := orderBook.AddOrder(order)
```

#### Market Data
```go
// Get order book depth
depth := orderBook.GetDepth(10)

// Get market summary
summary := orderBook.GetMarketSummary()

// Subscribe to updates
wsClient.Subscribe("orderbook", "BTC/USDT")
```

### Lending Protocol Integration

#### Pool Management
```go
// Create lending pool
pool := &lending.LendingPool{
    Asset:              "ETH",
    MaxSupply:          big.NewInt(100000000000000000000),
    InterestRate:       big.NewInt(500),
    LiquidationThreshold: big.NewInt(8000),
}

err := lendingProtocol.CreatePool(pool)
```

#### Supply/Borrow Operations
```go
// Supply assets
err := lendingProtocol.Supply(poolID, amount, user)

// Borrow assets
err := lendingProtocol.Borrow(poolID, amount, user)

// Get user position
position := lendingProtocol.GetUserPosition(user)
```

## Cross-Chain Bridge Development

### Bridge Architecture

```
Source Chain          Bridge Network          Destination Chain
     │                      │                       │
     │  Lock Assets         │                       │
     ├─────────────────────►│                       │
     │                      │  Validate & Confirm   │
     │                      ├─────────────────────►│
     │                      │                       │  Mint Assets
     │                      │                       ├─►
     │                      │  Confirmation         │
     │                      │◄──────────────────────┤
     │  Unlock Assets       │                       │
     │◄─────────────────────┤                       │
```

### Validator Management

```go
// Add validator
validator := &bridge.Validator{
    ID:          "validator_123",
    Address:     "0x...",
    StakeAmount: big.NewInt(1000000000000000000),
    IsActive:    true,
}

err := bridge.AddValidator(validator, stakeAmount, publicKey)

// Update stake
err := bridge.UpdateValidatorStake(validatorID, newStakeAmount)

// Remove validator
err := bridge.RemoveValidator(validatorID)
```

### Asset Mapping

```go
// Create asset mapping
mapping := &bridge.AssetMapping{
    SourceChain:      bridge.ChainIDGoChain,
    DestinationChain: bridge.ChainIDEthereum,
    SourceAsset:      "0x...",
    DestinationAsset: "0x...",
    AssetType:        bridge.AssetTypeERC20,
    Decimals:         18,
    MinAmount:        big.NewInt(1000000000000000),
    MaxAmount:        big.NewInt(1000000000000000000),
}

err := bridge.CreateAssetMapping(mapping)
```

## Governance & DAO Development

### Proposal System

#### Create Proposal
```go
proposal := &governance.Proposal{
    Title:          "Increase Treasury Limit",
    Description:    "Proposal to increase daily treasury spending limit",
    ProposalType:   governance.ProposalTypeTreasury,
    QuorumRequired: big.NewInt(1000000000000000000),
    MinVotingPower: big.NewInt(100000000000000000),
}

err := governanceSystem.CreateProposal(proposal)
```

#### Voting Mechanism
```go
// Cast vote
vote := &governance.Vote{
    ProposalID:  proposalID,
    Voter:       "0x...",
    VoteChoice:  governance.VoteChoiceFor,
    VotingPower: big.NewInt(100000000000000000),
    Reason:      "This proposal will improve efficiency",
}

err := governanceSystem.CastVote(proposalID, voter, voteChoice, reason)

// Delegate voting power
err := governanceSystem.DelegateVotingPower(delegator, delegate, amount)
```

### Treasury Management

#### Treasury Operations
```go
// Create treasury proposal
treasuryProposal := &governance.TreasuryProposal{
    Title:       "Fund Development",
    Description: "Allocate funds for development team",
    Amount:      big.NewInt(5000000000000000000),
    Asset:       "ETH",
    Recipient:   "0x...",
    Purpose:     "Development team expansion",
}

err := treasury.CreateProposal(treasuryProposal)

// Execute transaction
err := treasury.CreateDirectTransaction(
    governance.TreasuryTransactionTypeTransfer,
    amount,
    asset,
    recipient,
    description,
    executor,
)
```

## Testing & Deployment

### Unit Testing

#### Test Structure
```go
func TestOrderBook(t *testing.T) {
    // Setup
    orderBook, err := orderbook.NewOrderBook("BTC/USDT")
    require.NoError(t, err)
    
    // Test data
    order := &orderbook.Order{
        ID:          "test_order",
        TradingPair: "BTC/USDT",
        Side:        orderbook.OrderSideBuy,
        Type:        orderbook.OrderTypeLimit,
        Quantity:    big.NewInt(1000000),
        Price:       big.NewInt(50000),
    }
    
    // Execute
    err = orderBook.AddOrder(order)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, orderbook.OrderStatusPending, order.Status)
}
```

#### Integration Testing
```go
func TestTradingFlow(t *testing.T) {
    // Setup trading environment
    env, err := setupTradingEnvironment()
    require.NoError(t, err)
    
    // Create orders
    buyOrder := createTestOrder("buy", 50000, 1000000)
    sellOrder := createTestOrder("sell", 50000, 1000000)
    
    // Add orders
    err = env.OrderBook.AddOrder(buyOrder)
    require.NoError(t, err)
    
    err = env.OrderBook.AddOrder(sellOrder)
    require.NoError(t, err)
    
    // Verify matching
    trades := env.MatchingEngine.GetTrades()
    assert.Len(t, trades, 1)
    assert.Equal(t, big.NewInt(50000), trades[0].Price)
}
```

### Performance Testing

#### Benchmark Tests
```go
func BenchmarkOrderBookAddOrder(b *testing.B) {
    orderBook, _ := orderbook.NewOrderBook("BTC/USDT")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        order := &orderbook.Order{
            ID:          fmt.Sprintf("order_%d", i),
            TradingPair: "BTC/USDT",
            Side:        orderbook.OrderSideBuy,
            Type:        orderbook.OrderTypeLimit,
            Quantity:    big.NewInt(1000000),
            Price:       big.NewInt(50000 + i),
        }
        orderBook.AddOrder(order)
    }
}
```

### Deployment

#### Production Configuration
```yaml
# config/production.yaml
server:
  port: 443
  host: "0.0.0.0"
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/gochain.crt"
    key_file: "/etc/ssl/private/gochain.key"

blockchain:
  network_id: 1
  genesis_file: "config/mainnet-genesis.json"
  data_dir: "/var/lib/gochain"

exchange:
  enabled: true
  max_orders: 10000000
  rate_limit: 1000

bridge:
  enabled: true
  min_validators: 5
  security_level: "high"
```

#### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/node

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./main"]
```

## Performance Optimization

### Database Optimization

1. **Indexing Strategy**
   ```go
   // Create indexes for frequently queried fields
   db.CreateIndex("orders", "trading_pair", "side", "price")
   db.CreateIndex("transactions", "from", "to", "block_number")
   db.CreateIndex("proposals", "status", "created_at")
   ```

2. **Connection Pooling**
   ```go
   // Configure connection pool
   db.SetMaxOpenConns(100)
   db.SetMaxIdleConns(10)
   db.SetConnMaxLifetime(time.Hour)
   ```

### Caching Strategy

1. **In-Memory Cache**
   ```go
   // Use LRU cache for frequently accessed data
   cache := lru.New(1000)
   
   // Cache order book data
   cache.Add("orderbook_BTC_USDT", orderBookData)
   
   // Retrieve cached data
   if data, ok := cache.Get("orderbook_BTC_USDT"); ok {
       return data.(*OrderBookData)
   }
   ```

2. **Redis Cache**
   ```go
   // Configure Redis client
   redisClient := redis.NewClient(&redis.Options{
       Addr:     "localhost:6379",
       Password: "",
       DB:       0,
   })
   
   // Cache market data
   err := redisClient.Set(ctx, "market_data_BTC_USDT", data, time.Hour).Err()
   ```

### Concurrency Optimization

1. **Goroutine Pools**
   ```go
   // Create worker pool for order processing
   pool := workerpool.New(100)
   
   for i := 0; i < 1000; i++ {
       pool.Submit(func() {
           processOrder(orders[i])
       })
   }
   
   pool.StopWait()
   ```

2. **Channel Buffering**
   ```go
   // Buffer channels for better performance
   orderChannel := make(chan *Order, 10000)
   tradeChannel := make(chan *Trade, 1000)
   ```

## Security Best Practices

### Input Validation

1. **Sanitize User Inputs**
   ```go
   // Validate trading pair format
   func validateTradingPair(pair string) error {
       if !regexp.MustCompile(`^[A-Z]{3,4}/[A-Z]{3,4}$`).MatchString(pair) {
           return errors.New("invalid trading pair format")
       }
       return nil
   }
   
   // Validate amounts
   func validateAmount(amount *big.Int) error {
       if amount.Cmp(big.NewInt(0)) <= 0 {
           return errors.New("amount must be positive")
       }
       if amount.Cmp(maxAmount) > 0 {
           return errors.New("amount exceeds maximum")
       }
       return nil
   }
   ```

2. **SQL Injection Prevention**
   ```go
   // Use parameterized queries
   query := "SELECT * FROM orders WHERE trading_pair = ? AND user_id = ?"
   rows, err := db.Query(query, tradingPair, userID)
   
   // Avoid string concatenation
   // WRONG: query := "SELECT * FROM orders WHERE trading_pair = '" + tradingPair + "'"
   ```

### Access Control

1. **Role-Based Access Control**
   ```go
   type Role string
   
   const (
       RoleUser      Role = "user"
       RoleTrader    Role = "trader"
       RoleAdmin     Role = "admin"
       RoleValidator Role = "validator"
   )
   
   func (u *User) HasPermission(action string) bool {
       switch action {
       case "create_order":
           return u.Role == RoleTrader || u.Role == RoleAdmin
       case "manage_validators":
           return u.Role == RoleAdmin
       default:
           return false
       }
   }
   ```

2. **API Key Management**
   ```go
   // Validate API key
   func validateAPIKey(key string) (*APIKey, error) {
       if key == "" {
           return nil, errors.New("API key required")
       }
       
       apiKey, exists := apiKeyStore.Get(key)
       if !exists {
           return nil, errors.New("invalid API key")
       }
       
       if apiKey.ExpiresAt.Before(time.Now()) {
           return nil, errors.New("API key expired")
       }
       
       return apiKey, nil
   }
   ```

### Rate Limiting

1. **Implement Rate Limiting**
   ```go
   // Rate limiter per user
   rateLimiter := rate.NewLimiter(rate.Limit(100), 1000) // 100 req/sec, burst 1000
   
   func handleRequest(userID string) error {
       if !rateLimiter.Allow() {
           return errors.New("rate limit exceeded")
       }
       // Process request
       return nil
   }
   ```

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database status
   sudo systemctl status postgresql
   
   # Check connection
   psql -h localhost -U gochain -d gochain
   
   # Check logs
   tail -f /var/log/postgresql/postgresql-15-main.log
   ```

2. **Memory Issues**
   ```bash
   # Check memory usage
   free -h
   
   # Check Go memory stats
   curl http://localhost:8080/debug/pprof/heap
   
   # Profile memory usage
   go tool pprof http://localhost:8080/debug/pprof/heap
   ```

3. **Performance Issues**
   ```bash
   # Check CPU usage
   top
   
   # Check Go profiling
   curl http://localhost:8080/debug/pprof/profile
   
   # Profile CPU usage
   go tool pprof http://localhost:8080/debug/pprof/profile
   ```

### Debug Tools

1. **Go Debugging**
   ```go
   // Enable debug logging
   log.SetLevel(log.DebugLevel)
   
   // Add debug statements
   log.Debugf("Processing order: %+v", order)
   
   // Use pprof for profiling
   import _ "net/http/pprof"
   ```

2. **Monitoring**
   ```go
   // Add metrics
   import "github.com/prometheus/client_golang/prometheus"
   
   var (
       ordersProcessed = prometheus.NewCounter(prometheus.CounterOpts{
           Name: "orders_processed_total",
           Help: "Total number of orders processed",
       })
   )
   
   // Record metrics
   ordersProcessed.Inc()
   ```

## Conclusion

This developer guide provides a comprehensive overview of developing on the GoChain platform. For more detailed information, refer to the specific component documentation and API references.

### Next Steps

1. **Explore Examples**: Check out the `examples/` directory for working code samples
2. **Join Community**: Participate in discussions on GitHub and Discord
3. **Build Something**: Start with a simple DApp and gradually add complexity
4. **Contribute**: Submit issues, pull requests, and help improve the platform

### Resources

- **Documentation**: [docs.gochain.io](https://docs.gochain.io)
- **GitHub**: [github.com/gochain/gochain](https://github.com/gochain/gochain)
- **Discord**: [discord.gg/gochain](https://discord.gg/gochain)
- **Blog**: [blog.gochain.io](https://blog.gochain.io)

Happy coding on GoChain! 🚀
