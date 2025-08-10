# GoChain

A comprehensive, production-ready blockchain implementation written in Go. GoChain is a modular blockchain platform featuring proof-of-work consensus, P2P networking, transaction mempool, UTXO management, and secure wallet functionality with encryption.

## 🚀 Features

### Core Blockchain
- **Block Structure**: Complete block implementation with headers, transactions, and Merkle root validation
- **Proof-of-Work Consensus**: Configurable mining difficulty with automatic adjustment
- **Transaction Model**: UTXO-based transaction system with proper change calculation
- **Chain Management**: Genesis block creation, block validation, and difficulty calculation

### Security & Cryptography
- **ECDSA Signatures**: P-256 curve for transaction signing and verification
- **Address Generation**: SHA-256-based address derivation with checksum validation
- **Wallet Encryption**: AES-256 encryption with PBKDF2 key derivation
- **Canonical Signatures**: Low-S signature enforcement for transaction security

### Networking & Storage
- **P2P Networking**: libp2p-based peer-to-peer communication with GossipSub
- **Persistent Storage**: File-based storage system with encryption support
- **UTXO Management**: Complete unspent transaction output tracking
- **Mempool**: Transaction pool with fee-based eviction policies

### Wallet & Transactions
- **Multi-Account Support**: Create and manage multiple wallet accounts
- **Transaction Creation**: Automated UTXO selection and change calculation
- **Key Management**: Secure private key storage and import/export functionality
- **Balance Tracking**: Real-time balance updates and UTXO monitoring

## 🏗️ Architecture

```
GoChain/
├── cmd/gochain/          # Main CLI application
├── pkg/
│   ├── block/            # Block structure and validation
│   ├── chain/            # Blockchain management and consensus
│   ├── consensus/        # Proof-of-work mining
│   ├── mempool/          # Transaction pool management
│   ├── miner/            # Block mining and assembly
│   ├── net/              # P2P networking (libp2p)
│   ├── proto/            # Protocol buffer definitions
│   ├── storage/          # Persistent storage layer
│   ├── utxo/             # UTXO set management
│   └── wallet/           # Wallet and key management
└── config/               # Configuration files
```

### Component Overview

- **`cmd/gochain`**: Main CLI that orchestrates all components into a full node
- **`pkg/block`**: Block, header, and transaction structures with validation
- **`pkg/chain`**: Blockchain state management, genesis creation, and difficulty adjustment
- **`pkg/consensus`**: Proof-of-work mining and block validation
- **`pkg/mempool`**: Transaction pool with fee-based prioritization
- **`pkg/miner`**: Block assembly and mining loop
- **`pkg/net`**: P2P networking with peer discovery and message routing
- **`pkg/storage`**: Encrypted file-based storage system
- **`pkg/utxo`**: UTXO tracking, balance calculation, and double-spend prevention
- **`pkg/wallet`**: Key generation, transaction signing, and secure storage

## 📦 Installation

### Prerequisites
- Go 1.20 or later
- Git

### Quick Start
```bash
# Clone the repository
git clone https://github.com/gochain/gochain
cd gochain

# Build the full node
go build -o gochain ./cmd/gochain

# Run the node
./gochain --help
```

### Build Options
```bash
# Build with all features enabled
go build -o gochain ./cmd/gochain

# Build specific packages
go build ./pkg/block
go build ./pkg/wallet
```

## ⚙️ Configuration

### Configuration File
Create `config/config.yaml`:
```yaml
network:
  port: 30303
  bootstrap_peers: []
  enable_mdns: true
  max_peers: 50

blockchain:
  genesis_reward: 1000000
  target_block_time: 10
  difficulty_adjustment_blocks: 2016

mining:
  enabled: false
  threads: 4
  coinbase_address: ""

mempool:
  max_size: 10000
  min_fee_rate: 1

wallet:
  key_type: "secp256r1"
  passphrase: ""

storage:
  data_dir: "./data"
  db_type: "file"
```

### Environment Variables
- `GOCHAIN_CONFIG`: Path to configuration file
- `GOCHAIN_NETWORK`: Network type (mainnet/testnet/devnet)
- `GOCHAIN_PORT`: Network port (0 for random)
- `GOCHAIN_WALLET_FILE`: Wallet file path
- `GOCHAIN_PASSPHRASE`: Wallet encryption passphrase

## 🖥️ Usage

### Running a Node
```bash
# Start a full node with mining enabled
./gochain --mining --port 30303

# Start with custom configuration
./gochain --config ./config/config.yaml --mining

# Start in testnet mode
./gochain --network testnet --port 0
```

### Wallet Management
```bash
# Create a new wallet
./gochain wallet --wallet-file mywallet.dat --passphrase "secure_passphrase"

# Check wallet balance
./gochain balance --address <address> --wallet-file mywallet.dat --passphrase "secure_passphrase"

# Send a transaction
./gochain send --from <from_address> --to <to_address> --amount 1000 --fee 10 --wallet-file mywallet.dat --passphrase "secure_passphrase"
```

### Blockchain Information
```bash
# Get chain status
./gochain info

# View block details
./gochain block <block_hash>

# Check transaction status
./gochain tx <tx_hash>
```

## 🧪 Testing

### Run All Tests
```bash
# Test all packages
go test ./...

# Test specific package
go test ./pkg/wallet -v

# Test with coverage
go test -cover ./pkg/...
```

### Test Results
All tests are currently passing:
- ✅ Block validation and creation
- ✅ Chain management and consensus
- ✅ Wallet functionality and encryption
- ✅ UTXO management and validation
- ✅ Network connectivity and messaging
- ✅ Storage persistence and encryption
- ✅ Transaction creation and signing

## 🔒 Security Features

### Cryptographic Security
- **ECDSA P-256**: Industry-standard elliptic curve cryptography
- **AES-256-GCM**: Authenticated encryption for wallet data
- **PBKDF2**: Key derivation with configurable iterations
- **SHA-256**: Secure hashing for addresses and blocks

### Wallet Security
- **Encrypted Storage**: All private keys are encrypted on disk
- **Passphrase Protection**: User-defined encryption keys
- **Secure Key Generation**: Cryptographically secure random number generation
- **Import/Export Security**: Safe private key transfer mechanisms

### Network Security
- **Message Signing**: All network messages are cryptographically signed
- **Peer Validation**: Secure peer discovery and connection validation
- **DoS Protection**: Rate limiting and connection management

## 🚧 Development Status

### ✅ Completed Features
- [x] Complete blockchain implementation
- [x] Proof-of-work consensus mechanism
- [x] UTXO-based transaction model
- [x] Secure wallet with encryption
- [x] P2P networking infrastructure
- [x] Persistent storage system
- [x] Comprehensive test coverage
- [x] CLI interface and commands

### 🔄 In Progress
- [ ] Smart contract support
- [ ] Advanced consensus mechanisms
- [ ] Enhanced P2P protocols
- [ ] API and RPC endpoints

### 📋 Planned Features
- [ ] Layer 2 scaling solutions
- [ ] Cross-chain interoperability
- [ ] Advanced privacy features
- [ ] Governance mechanisms

## 🤝 Contributing

### Development Setup
```bash
# Fork and clone
git clone https://github.com/yourusername/gochain
cd gochain

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./cmd/gochain
```

### Code Standards
- Follow Go formatting standards (`gofmt`)
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting PRs

### Testing Guidelines
- Unit tests for all new functionality
- Integration tests for complex features
- Performance benchmarks for critical paths
- Security tests for cryptographic functions

## 📚 Documentation

### API Reference
- [Block Package](pkg/block/README.md)
- [Chain Package](pkg/chain/README.md)
- [Wallet Package](pkg/wallet/README.md)
- [Network Package](pkg/net/README.md)

### Examples
- [Basic Usage](examples/basic_usage.go)
- [Wallet Operations](examples/wallet_operations.go)
- [Network Setup](examples/network_setup.go)

## 🔍 Troubleshooting

### Common Issues

**Build Errors**
```bash
# Ensure Go version is 1.20+
go version

# Clean and rebuild
go clean -cache
go build ./...
```

**Network Issues**
```bash
# Check port availability
netstat -tulpn | grep :30303

# Verify firewall settings
sudo ufw status
```

**Wallet Issues**
```bash
# Verify wallet file permissions
ls -la wallet.dat

# Check passphrase correctness
./gochain wallet --help
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **libp2p**: P2P networking infrastructure
- **Go Standard Library**: Cryptographic primitives
- **Protocol Buffers**: Message serialization
- **BadgerDB**: Storage backend (optional)

---

**GoChain** - Building the future of decentralized applications, one block at a time. 🚀
