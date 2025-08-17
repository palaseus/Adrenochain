# Smart Contract Development Guide 🚀

This guide covers smart contract development on adrenochain, including EVM and WASM execution environments, contract development best practices, and advanced features.

## 🎯 **Overview**

adrenochain provides a comprehensive smart contract platform with support for:

- **Ethereum Virtual Machine (EVM)**: Full Ethereum compatibility
- **WebAssembly (WASM)**: Cross-platform contract execution
- **Unified Development Interface**: Consistent API for both engines
- **Advanced Features**: Gas optimization, state management, and security tools

## 🏗️ **Architecture**

### **Smart Contract Engine Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│  Contract Deployment  │  Contract Interaction  │  Events   │
│  [Deploy]            │  [Call/Execute]        │  [Logs]   │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   Execution Layer                           │
├─────────────────────────────────────────────────────────────┤
│  EVM Engine         │  WASM Engine          │  Gas Meter │
│  [Solidity/Vyper]  │  [Rust/AssemblyScript]│  [Tracking]│
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   State Management                          │
├─────────────────────────────────────────────────────────────┤
│  Contract Storage   │  State Trie           │  Persistence│
│  [Key-Value]       │  [Merkle Patricia]    │  [LevelDB]  │
└─────────────────────────────────────────────────────────────┘
```

### **Key Components**

1. **Contract Engine**: Manages contract execution and lifecycle
2. **State Manager**: Handles contract state persistence
3. **Gas Meter**: Tracks and limits execution costs
4. **Event System**: Manages contract events and logs
5. **Security Layer**: Provides access control and validation

## 🔧 **EVM Development**

### **Supported Languages**

- **Solidity**: Primary smart contract language
- **Vyper**: Python-like alternative
- **Yul**: Low-level assembly language
- **Yul+**: Enhanced Yul with additional features

### **Solidity Contract Example**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract adrenochainToken is ERC20, Ownable {
    uint256 public constant INITIAL_SUPPLY = 1000000 * 10**18;
    
    constructor() ERC20("adrenochain Token", "GCH") {
        _mint(msg.sender, INITIAL_SUPPLY);
    }
    
    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }
    
    function burn(uint256 amount) public {
        _burn(msg.sender, amount);
    }
}
```

### **Contract Deployment**

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/adrenochain/adrenochain/pkg/contracts/evm"
    "github.com/adrenochain/adrenochain/pkg/wallet"
)

func main() {
    // Initialize EVM engine
    engine := evm.NewEngine()
    
    // Load wallet
    wallet, err := wallet.LoadWallet("wallet.dat", "password")
    if err != nil {
        log.Fatal(err)
    }
    
    // Compile contract (assuming you have the bytecode)
    bytecode := []byte("...") // Contract bytecode
    
    // Deploy contract
    contract, err := engine.DeployContract(context.Background(), &evm.DeployRequest{
        Bytecode: bytecode,
        GasLimit: 5000000,
        Value:    0,
        Wallet:   wallet,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Contract deployed at: %s\n", contract.Address)
}
```

### **Contract Interaction**

```go
// Call contract function
result, err := engine.CallContract(context.Background(), &evm.CallRequest{
    ContractAddress: contract.Address,
    Function:        "balanceOf",
    Arguments:       []interface{}{wallet.Address()},
    GasLimit:        100000,
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Balance: %s\n", result.ReturnData)

// Execute contract function
tx, err := engine.ExecuteContract(context.Background(), &evm.ExecuteRequest{
    ContractAddress: contract.Address,
    Function:        "transfer",
    Arguments:       []interface{}{recipient, amount},
    GasLimit:        100000,
    Wallet:          wallet,
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Transaction hash: %s\n", tx.Hash)
```

## 🌐 **WASM Development**

### **Supported Languages**

- **Rust**: Primary WASM language with excellent tooling
- **AssemblyScript**: TypeScript-like language for WASM
- **C/C++**: Traditional languages with WASM compilation
- **Go**: Experimental Go to WASM compilation

### **Rust Contract Example**

```rust
use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct Token {
    pub name: String,
    pub symbol: String,
    pub total_supply: u64,
    pub balances: std::collections::HashMap<String, u64>,
}

#[wasm_bindgen]
impl Token {
    pub fn new(name: String, symbol: String, initial_supply: u64) -> Token {
        let mut balances = std::collections::HashMap::new();
        balances.insert("owner".to_string(), initial_supply);
        
        Token {
            name,
            symbol,
            total_supply: initial_supply,
            balances,
        }
    }
    
    pub fn transfer(&mut self, from: String, to: String, amount: u64) -> bool {
        if let Some(balance) = self.balances.get_mut(&from) {
            if *balance >= amount {
                *balance -= amount;
                *self.balances.entry(to).or_insert(0) += amount;
                return true;
            }
        }
        false
    }
    
    pub fn balance_of(&self, address: String) -> u64 {
        *self.balances.get(&address).unwrap_or(&0)
    }
    
    pub fn get_name(&self) -> String {
        self.name.clone()
    }
    
    pub fn get_symbol(&self) -> String {
        self.symbol.clone()
    }
    
    pub fn get_total_supply(&self) -> u64 {
        self.total_supply
    }
}
```

### **WASM Contract Deployment**

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/adrenochain/adrenochain/pkg/contracts/wasm"
    "github.com/adrenochain/adrenochain/pkg/wallet"
)

func main() {
    // Initialize WASM engine
    engine := wasm.NewEngine()
    
    // Load wallet
    wallet, err := wallet.LoadWallet("wallet.dat", "password")
    if err != nil {
        log.Fatal(err)
    }
    
    // Load WASM bytecode
    wasmBytes, err := os.ReadFile("token.wasm")
    if err != nil {
        log.Fatal(err)
    }
    
    // Deploy WASM contract
    contract, err := engine.DeployContract(context.Background(), &wasm.DeployRequest{
        Bytecode: wasmBytes,
        GasLimit: 1000000,
        Value:    0,
        Wallet:   wallet,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("WASM contract deployed at: %s\n", contract.Address)
}
```

## 🔒 **Security Best Practices**

### **1. Access Control**

```solidity
// Use OpenZeppelin's access control
import "@openzeppelin/contracts/access/AccessControl.sol";

contract SecureContract is AccessControl {
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");
    
    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
    }
    
    modifier onlyAdmin() {
        require(hasRole(ADMIN_ROLE, msg.sender), "Admin access required");
        _;
    }
    
    modifier onlyOperator() {
        require(hasRole(OPERATOR_ROLE, msg.sender), "Operator access required");
        _;
    }
}
```

### **2. Reentrancy Protection**

```solidity
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract SecureContract is ReentrancyGuard {
    mapping(address => uint256) private balances;
    
    function withdraw(uint256 amount) external nonReentrant {
        uint256 balance = balances[msg.sender];
        require(balance >= amount, "Insufficient balance");
        
        balances[msg.sender] = 0;
        
        (bool success, ) = msg.sender.call{value: amount}("");
        require(success, "Transfer failed");
    }
}
```

### **3. Integer Overflow Protection**

```solidity
// Use SafeMath or Solidity 0.8.0+
import "@openzeppelin/contracts/utils/math/SafeMath.sol";

contract SafeContract {
    using SafeMath for uint256;
    
    function safeAdd(uint256 a, uint256 b) public pure returns (uint256) {
        return a.add(b);
    }
    
    function safeSub(uint256 a, uint256 b) public pure returns (uint256) {
        return a.sub(b);
    }
}
```

### **4. Input Validation**

```solidity
contract ValidatedContract {
    function processData(bytes calldata data) external {
        require(data.length > 0, "Data cannot be empty");
        require(data.length <= 1024, "Data too large");
        
        // Process validated data
        _processData(data);
    }
    
    function _processData(bytes calldata data) internal {
        // Implementation
    }
}
```

## ⛽ **Gas Optimization**

### **1. Storage Optimization**

```solidity
contract GasOptimized {
    // Pack related data into single storage slots
    struct User {
        uint128 balance;      // 16 bytes
        uint64 lastUpdate;    // 8 bytes
        uint64 userId;        // 8 bytes
    } // Total: 32 bytes (1 storage slot)
    
    // Use uint256 for single values to avoid packing overhead
    uint256 public totalSupply;
    
    // Use mappings instead of arrays for large datasets
    mapping(address => User) public users;
}
```

### **2. Function Optimization**

```solidity
contract OptimizedFunctions {
    // Use external for functions only called externally
    function externalFunction() external pure returns (uint256) {
        return 42;
    }
    
    // Use view/pure when possible
    function getBalance(address user) external view returns (uint256) {
        return balances[user];
    }
    
    // Batch operations to reduce gas costs
    function batchTransfer(address[] calldata recipients, uint256[] calldata amounts) external {
        require(recipients.length == amounts.length, "Length mismatch");
        
        for (uint256 i = 0; i < recipients.length; i++) {
            _transfer(msg.sender, recipients[i], amounts[i]);
        }
    }
}
```

### **3. Event Optimization**

```solidity
contract EventOptimized {
    // Use indexed parameters for efficient filtering
    event Transfer(
        address indexed from,
        address indexed to,
        uint256 indexed tokenId
    );
    
    // Use custom events for complex data
    event UserUpdated(
        address indexed user,
        uint256 timestamp,
        string action
    );
}
```

## 🧪 **Testing Smart Contracts**

### **1. Unit Testing with Go**

```go
package contracts_test

import (
    "testing"
    
    "github.com/adrenochain/adrenochain/pkg/contracts/evm"
    "github.com/stretchr/testify/assert"
)

func TestTokenContract(t *testing.T) {
    // Setup test environment
    engine := evm.NewTestEngine()
    
    // Deploy test contract
    contract, err := engine.DeployTestContract("Token.sol")
    assert.NoError(t, err)
    
    // Test contract functions
    t.Run("Initial Supply", func(t *testing.T) {
        result, err := engine.CallTestFunction(contract, "totalSupply")
        assert.NoError(t, err)
        assert.Equal(t, "1000000", result.ReturnData)
    })
    
    t.Run("Transfer", func(t *testing.T) {
        // Test transfer functionality
        err := engine.ExecuteTestFunction(contract, "transfer", "recipient", "100")
        assert.NoError(t, err)
        
        // Verify balance
        balance, err := engine.CallTestFunction(contract, "balanceOf", "recipient")
        assert.NoError(t, err)
        assert.Equal(t, "100", balance.ReturnData)
    })
}
```

### **2. Integration Testing**

```go
func TestContractIntegration(t *testing.T) {
    // Setup full blockchain environment
    chain := setupTestChain(t)
    engine := evm.NewEngine()
    
    // Deploy multiple contracts
    token := deployTokenContract(t, engine, chain)
    amm := deployAMMContract(t, engine, chain)
    
    // Test contract interaction
    t.Run("AMM Integration", func(t *testing.T) {
        // Test AMM with token contract
        err := amm.AddLiquidity(token.Address, "1000", "1000")
        assert.NoError(t, err)
        
        // Verify liquidity addition
        liquidity := amm.GetLiquidity(token.Address)
        assert.Equal(t, "1000", liquidity)
    })
}
```

### **3. Fuzz Testing**

```go
func FuzzTokenTransfer(f *testing.F) {
    // Add seed corpus
    f.Add("sender", "recipient", uint64(100))
    f.Add("sender", "recipient", uint64(1000000))
    
    f.Fuzz(func(t *testing.T, sender, recipient string, amount uint64) {
        // Setup test environment
        engine := evm.NewTestEngine()
        contract := deployTestToken(t, engine)
        
        // Fuzz transfer parameters
        err := engine.ExecuteTestFunction(contract, "transfer", sender, recipient, amount)
        
        // Verify no panics or unexpected errors
        if err != nil {
            // Log error for analysis
            t.Logf("Transfer failed: %v", err)
        }
    })
}
```

## 📊 **Contract Analytics**

### **1. Gas Usage Analysis**

```go
package analytics

import (
    "github.com/adrenochain/adrenochain/pkg/contracts/evm"
)

type GasAnalytics struct {
    engine *evm.Engine
}

func (ga *GasAnalytics) AnalyzeGasUsage(contractAddress string) (*GasReport, error) {
    // Analyze contract gas usage patterns
    report := &GasReport{
        ContractAddress: contractAddress,
        Functions:       make(map[string]*FunctionGas),
    }
    
    // Get contract functions
    functions, err := ga.engine.GetContractFunctions(contractAddress)
    if err != nil {
        return nil, err
    }
    
    // Analyze each function
    for _, function := range functions {
        gasUsage, err := ga.analyzeFunctionGas(contractAddress, function)
        if err != nil {
            continue
        }
        
        report.Functions[function] = gasUsage
    }
    
    return report, nil
}

type GasReport struct {
    ContractAddress string                    `json:"contract_address"`
    Functions       map[string]*FunctionGas  `json:"functions"`
    TotalGas        uint64                   `json:"total_gas"`
}

type FunctionGas struct {
    Name           string  `json:"name"`
    AverageGas     uint64  `json:"average_gas"`
    MinGas         uint64  `json:"min_gas"`
    MaxGas         uint64  `json:"max_gas"`
    CallCount      uint64  `json:"call_count"`
}
```

### **2. Contract Performance Monitoring**

```go
type ContractMonitor struct {
    engine *evm.Engine
    metrics map[string]*ContractMetrics
}

func (cm *ContractMonitor) MonitorContract(contractAddress string) {
    // Monitor contract performance in real-time
    go func() {
        for {
            // Collect metrics
            metrics := cm.collectMetrics(contractAddress)
            
            // Update monitoring data
            cm.metrics[contractAddress] = metrics
            
            // Wait for next collection cycle
            time.Sleep(30 * time.Second)
        }
    }()
}

func (cm *ContractMonitor) collectMetrics(contractAddress string) *ContractMetrics {
    // Collect various contract metrics
    return &ContractMetrics{
        Timestamp:     time.Now(),
        GasUsed:       cm.getGasUsage(contractAddress),
        CallCount:     cm.getCallCount(contractAddress),
        ErrorRate:     cm.getErrorRate(contractAddress),
        ResponseTime:  cm.getResponseTime(contractAddress),
    }
}
```

## 🚀 **Advanced Features**

### **1. Contract Upgradeability**

```solidity
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Upgrade.sol";

contract UpgradeableContract is ERC1967Upgrade {
    // Implementation can be upgraded
    function upgradeTo(address newImplementation) external onlyOwner {
        _upgradeToAndCall(newImplementation, "", false);
    }
}
```

### **2. Meta-Transactions**

```solidity
contract MetaTransactionContract {
    struct MetaTransaction {
        address from;
        uint256 nonce;
        bytes functionSignature;
    }
    
    mapping(address => uint256) public nonces;
    
    function executeMetaTransaction(
        address userAddress,
        bytes memory functionSignature,
        bytes32 sigR,
        bytes32 sigS,
        uint8 sigV
    ) public returns (bytes memory) {
        MetaTransaction memory metaTx = MetaTransaction({
            from: userAddress,
            nonce: nonces[userAddress]++,
            functionSignature: functionSignature
        });
        
        require(verify(userAddress, metaTx, sigR, sigS, sigV), "Invalid signature");
        
        // Execute the function
        (bool success, bytes memory returnData) = address(this).call(
            abi.encodePacked(functionSignature, userAddress)
        );
        
        require(success, "Function call failed");
        return returnData;
    }
}
```

### **3. Contract Factory Pattern**

```solidity
contract ContractFactory {
    event ContractCreated(address indexed contract, address indexed owner);
    
    function createContract(bytes memory bytecode, bytes memory constructorArgs) 
        external 
        returns (address contractAddress) 
    {
        // Create contract with constructor arguments
        assembly {
            contractAddress := create2(
                0,
                add(bytecode, 0x20),
                mload(bytecode),
                keccak256(constructorArgs, mload(constructorArgs))
            )
        }
        
        require(contractAddress != address(0), "Contract creation failed");
        
        emit ContractCreated(contractAddress, msg.sender);
        return contractAddress;
    }
}
```

## 📚 **Development Tools**

### **1. adrenochain CLI**

```bash
# Deploy contract
adrenochain contract deploy Token.sol --network testnet

# Call contract function
adrenochain contract call 0x123... balanceOf 0x456...

# Execute contract function
adrenochain contract execute 0x123... transfer 0x456... 100

# Verify contract
adrenochain contract verify 0x123... Token.sol --constructor-args "Token,TKN,1000000"
```

### **2. Development Environment**

```bash
# Install development dependencies
go mod download

# Run tests
go test ./pkg/contracts/... -v

# Run benchmarks
go test ./pkg/contracts/... -bench=.

# Generate documentation
godoc -http=:6060
```

### **3. IDE Integration**

- **VS Code**: adrenochain extension for contract development
- **GoLand**: Native Go support with contract debugging
- **Remix**: Web-based Solidity IDE integration

## 🔮 **Future Features**

### **1. Planned Enhancements**

- **Layer 2 Support**: Rollups and state channels
- **Cross-Chain Contracts**: Interoperability protocols
- **Advanced Privacy**: Zero-knowledge proof integration
- **AI Integration**: Machine learning in smart contracts

### **2. Research Areas**

- **Formal Verification**: Mathematical contract correctness proofs
- **Quantum Resistance**: Post-quantum cryptography support
- **Scalability**: Sharding and parallel execution
- **Governance**: DAO frameworks and voting systems

## 📞 **Support & Resources**

- **Documentation**: [docs.adrenochain.dev](https://docs.adrenochain.dev)
- **GitHub**: [github.com/adrenochain/adrenochain](https://github.com/adrenochain/adrenochain)
- **Discord**: [discord.gg/adrenochain](https://discord.gg/adrenochain)
- **Email**: contracts@adrenochain.dev

---

**Last Updated**: December 2024  
**Version**: 1.0.0  
**adrenochain**: Advanced smart contract development platform 🚀🔬💻
