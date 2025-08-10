# GoChain API Documentation

## Overview

GoChain is a blockchain implementation in Go that provides a complete blockchain infrastructure including wallet management, transaction processing, consensus mechanisms, and network communication.

## Package Structure

```
pkg/
├── block/      # Block and transaction structures
├── chain/      # Blockchain management and validation
├── consensus/  # Consensus mechanisms (PoW)
├── mempool/    # Transaction memory pool
├── miner/      # Block mining operations
├── net/        # Network communication
├── storage/    # Data persistence
├── utxo/       # Unspent Transaction Output management
└── wallet/     # Wallet operations and key management
```

## Core Types

### Block Structure

```go
type Block struct {
    Header       *BlockHeader
    Transactions []*Transaction
    Hash         []byte
    Size         int
}

type BlockHeader struct {
    Version       uint32
    PrevHash      []byte
    MerkleRoot    []byte
    Timestamp     int64
    Difficulty    uint64
    Nonce         uint64
}
```

### Transaction Structure

```go
type Transaction struct {
    ID        []byte
    Inputs    []*TxInput
    Outputs   []*TxOutput
    Timestamp int64
    Fee       uint64
}

type TxInput struct {
    TxID      []byte
    OutIndex  uint32
    Signature []byte
    PubKey    []byte
}

type TxOutput struct {
    Value   uint64
    Address string
}
```

## Package: block

### Block Operations

#### `NewBlock(prevHash []byte, transactions []*Transaction, difficulty uint64) *Block`
Creates a new block with the given parameters.

**Parameters:**
- `prevHash`: Hash of the previous block
- `transactions`: List of transactions to include
- `difficulty`: Mining difficulty target

**Returns:** New block instance

**Example:**
```go
block := block.NewBlock(prevHash, transactions, 1000000)
```

#### `(b *Block) CalculateHash() []byte`
Calculates the SHA-256 hash of the block.

**Returns:** Block hash as byte slice

**Example:**
```go
hash := block.CalculateHash()
```

#### `(b *Block) Validate() error`
Validates the block structure and contents.

**Returns:** Error if validation fails, nil otherwise

**Example:**
```go
if err := block.Validate(); err != nil {
    log.Printf("Block validation failed: %v", err)
}
```

### Transaction Operations

#### `NewTransaction(inputs []*TxInput, outputs []*TxOutput, fee uint64) *Transaction`
Creates a new transaction.

**Parameters:**
- `inputs`: Transaction inputs (UTXOs being spent)
- `outputs`: Transaction outputs (new UTXOs)
- `fee`: Transaction fee

**Returns:** New transaction instance

**Example:**
```go
tx := block.NewTransaction(inputs, outputs, 1000)
```

#### `(tx *Transaction) CalculateHash() []byte`
Calculates the transaction ID.

**Returns:** Transaction hash as byte slice

**Example:**
```go
txID := tx.CalculateHash()
```

#### `(tx *Transaction) Validate() error`
Validates the transaction structure and signatures.

**Returns:** Error if validation fails, nil otherwise

**Example:**
```go
if err := tx.Validate(); err != nil {
    log.Printf("Transaction validation failed: %v", err)
}
```

## Package: chain

### Blockchain Management

#### `NewChain(dataDir string) (*Chain, error)`
Creates a new blockchain instance.

**Parameters:**
- `dataDir`: Directory for storing blockchain data

**Returns:** Chain instance and error

**Example:**
```go
chain, err := chain.NewChain("./data")
if err != nil {
    log.Fatalf("Failed to create chain: %v", err)
}
```

#### `(c *Chain) AddBlock(block *block.Block) error`
Adds a new block to the blockchain.

**Parameters:**
- `block`: Block to add

**Returns:** Error if addition fails

**Example:**
```go
if err := chain.AddBlock(newBlock); err != nil {
    log.Printf("Failed to add block: %v", err)
}
```

#### `(c *Chain) GetBlock(hash []byte) (*block.Block, error)`
Retrieves a block by its hash.

**Parameters:**
- `hash`: Block hash

**Returns:** Block instance and error

**Example:**
```go
block, err := chain.GetBlock(blockHash)
if err != nil {
    log.Printf("Block not found: %v", err)
}
```

#### `(c *Chain) GetLatestBlock() *block.Block`
Gets the most recent block in the chain.

**Returns:** Latest block instance

**Example:**
```go
latestBlock := chain.GetLatestBlock()
log.Printf("Latest block height: %d", latestBlock.Header.Height)
```

#### `(c *Chain) ValidateBlock(block *block.Block) error`
Validates a block before adding it to the chain.

**Parameters:**
- `block`: Block to validate

**Returns:** Error if validation fails

**Example:**
```go
if err := chain.ValidateBlock(newBlock); err != nil {
    log.Printf("Block validation failed: %v", err)
}
```

#### `(c *Chain) GetBalance(address string) uint64`
Gets the balance of a specific address.

**Parameters:**
- `address`: Wallet address

**Returns:** Current balance in satoshis

**Example:**
```go
balance := chain.GetBalance("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")
log.Printf("Balance: %d satoshis", balance)
```

## Package: wallet

### Wallet Management

#### `NewWallet() (*Wallet, error)`
Creates a new wallet instance.

**Returns:** Wallet instance and error

**Example:**
```go
wallet, err := wallet.NewWallet()
if err != nil {
    log.Fatalf("Failed to create wallet: %v", err)
}
```

#### `(w *Wallet) GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error)`
Generates a new ECDSA key pair using secp256k1 curve.

**Returns:** Private key, public key, and error

**Example:**
```go
privKey, pubKey, err := wallet.GenerateKeyPair()
if err != nil {
    log.Printf("Failed to generate key pair: %v", err)
}
```

#### `(w *Wallet) GetAddress(pubKey *ecdsa.PublicKey) (string, error)`
Generates a Bitcoin-style address from a public key.

**Parameters:**
- `pubKey`: ECDSA public key

**Returns:** Base58-encoded address and error

**Example:**
```go
address, err := wallet.GetAddress(pubKey)
if err != nil {
    log.Printf("Failed to generate address: %v", err)
}
log.Printf("Generated address: %s", address)
```

#### `(w *Wallet) SignTransaction(tx *block.Transaction, privKey *ecdsa.PrivateKey) error`
Signs a transaction with the provided private key.

**Parameters:**
- `tx`: Transaction to sign
- `privKey`: Private key for signing

**Returns:** Error if signing fails

**Example:**
```go
if err := wallet.SignTransaction(transaction, privateKey); err != nil {
    log.Printf("Failed to sign transaction: %v", err)
}
```

#### `(w *Wallet) CreateTransaction(fromAddress, toAddress string, amount, fee uint64) (*block.Transaction, error)`
Creates a new transaction between two addresses.

**Parameters:**
- `fromAddress`: Source address
- `toAddress`: Destination address
- `amount`: Amount to transfer
- `fee`: Transaction fee

**Returns:** Created transaction and error

**Example:**
```go
tx, err := wallet.CreateTransaction(fromAddr, toAddr, 1000000, 1000)
if err != nil {
    log.Printf("Failed to create transaction: %v", err)
}
```

#### `(w *Wallet) VerifySignature(tx *block.Transaction, pubKey *ecdsa.PublicKey) bool`
Verifies the signature of a transaction.

**Parameters:**
- `tx`: Transaction to verify
- `pubKey`: Public key for verification

**Returns:** True if signature is valid

**Example:**
```go
if wallet.VerifySignature(transaction, publicKey) {
    log.Println("Transaction signature is valid")
} else {
    log.Println("Transaction signature is invalid")
}
```

## Package: utxo

### UTXO Management

#### `NewUTXOSet() *UTXOSet`
Creates a new UTXO set instance.

**Returns:** UTXO set instance

**Example:**
```go
utxoSet := utxo.NewUTXOSet()
```

#### `(us *UTXOSet) AddUTXO(utxo *UTXO)`
Adds a new UTXO to the set.

**Parameters:**
- `utxo`: UTXO to add

**Example:**
```go
utxoSet.AddUTXO(newUTXO)
```

#### `(us *UTXOSet) SpendUTXO(txID []byte, outIndex uint32) error`
Marks a UTXO as spent.

**Parameters:**
- `txID`: Transaction ID
- `outIndex`: Output index

**Returns:** Error if UTXO not found

**Example:**
```go
if err := utxoSet.SpendUTXO(txHash, 0); err != nil {
    log.Printf("Failed to spend UTXO: %v", err)
}
```

#### `(us *UTXOSet) GetAddressUTXOs(address string) []*UTXO`
Gets all unspent UTXOs for a specific address.

**Parameters:**
- `address`: Wallet address

**Returns:** List of unspent UTXOs

**Example:**
```go
utxos := utxoSet.GetAddressUTXOs("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")
log.Printf("Found %d unspent UTXOs", len(utxos))
```

#### `(us *UTXOSet) GetBalance(address string) uint64`
Calculates the total balance for an address.

**Parameters:**
- `address`: Wallet address

**Returns:** Total balance in satoshis

**Example:**
```go
balance := utxoSet.GetBalance("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")
log.Printf("Address balance: %d satoshis", balance)
```

## Package: miner

### Mining Operations

#### `NewMiner(chain *chain.Chain) *Miner`
Creates a new miner instance.

**Parameters:**
- `chain`: Blockchain instance

**Returns:** Miner instance

**Example:**
```go
miner := miner.NewMiner(blockchain)
```

#### `(m *Miner) MineBlock(transactions []*block.Transaction) (*block.Block, error)`
Mines a new block with the given transactions.

**Parameters:**
- `transactions`: Transactions to include in the block

**Returns:** Mined block and error

**Example:**
```go
block, err := miner.MineBlock(pendingTransactions)
if err != nil {
    log.Printf("Mining failed: %v", err)
} else {
    log.Printf("Block mined successfully: %x", block.Hash)
}
```

#### `(m *Miner) SetDifficulty(difficulty uint64)`
Sets the mining difficulty target.

**Parameters:**
- `difficulty`: New difficulty value

**Example:**
```go
miner.SetDifficulty(1000000)
```

#### `(m *Miner) GetHashRate() uint64`
Gets the current hash rate in hashes per second.

**Returns:** Hash rate

**Example:**
```go
hashRate := miner.GetHashRate()
log.Printf("Current hash rate: %d H/s", hashRate)
```

## Package: net

### Network Communication

#### `NewNetwork(port int, chain *chain.Chain) *Network`
Creates a new network instance.

**Parameters:**
- `port`: Network port to listen on
- `chain`: Blockchain instance

**Returns:** Network instance

**Example:**
```go
network := net.NewNetwork(8333, blockchain)
```

#### `(n *Network) Start() error`
Starts the network server.

**Returns:** Error if startup fails

**Example:**
```go
if err := network.Start(); err != nil {
    log.Fatalf("Failed to start network: %v", err)
}
```

#### `(n *Network) Stop()`
Stops the network server.

**Example:**
```go
network.Stop()
```

#### `(n *Network) Connect(peer string) error`
Connects to a peer node.

**Parameters:**
- `peer`: Peer address (host:port)

**Returns:** Error if connection fails

**Example:**
```go
if err := network.Connect("192.168.1.100:8333"); err != nil {
    log.Printf("Failed to connect to peer: %v", err)
}
```

#### `(n *Network) BroadcastTransaction(tx *block.Transaction) error`
Broadcasts a transaction to all connected peers.

**Parameters:**
- `tx`: Transaction to broadcast

**Returns:** Error if broadcast fails

**Example:**
```go
if err := network.BroadcastTransaction(transaction); err != nil {
    log.Printf("Failed to broadcast transaction: %v", err)
}
```

#### `(n *Network) BroadcastBlock(block *block.Block) error`
Broadcasts a new block to all connected peers.

**Parameters:**
- `block`: Block to broadcast

**Returns:** Error if broadcast fails

**Example:**
```go
if err := network.BroadcastBlock(newBlock); err != nil {
    log.Printf("Failed to broadcast block: %v", err)
}
```

## Package: storage

### Data Persistence

#### `NewStorage(dataDir string) (*Storage, error)`
Creates a new storage instance.

**Parameters:**
- `dataDir`: Directory for data storage

**Returns:** Storage instance and error

**Example:**
```go
storage, err := storage.NewStorage("./data")
if err != nil {
    log.Fatalf("Failed to create storage: %v", err)
}
```

#### `(s *Storage) Put(key []byte, value []byte) error`
Stores a key-value pair.

**Parameters:**
- `key`: Storage key
- `value`: Data to store

**Returns:** Error if storage fails

**Example:**
```go
if err := storage.Put([]byte("block:123"), blockData); err != nil {
    log.Printf("Failed to store block: %v", err)
}
```

#### `(s *Storage) Get(key []byte) ([]byte, error)`
Retrieves data by key.

**Parameters:**
- `key`: Storage key

**Returns:** Stored data and error

**Example:**
```go
data, err := storage.Get([]byte("block:123"))
if err != nil {
    log.Printf("Failed to retrieve block: %v", err)
}
```

#### `(s *Storage) Delete(key []byte) error`
Deletes data by key.

**Parameters:**
- `key`: Storage key

**Returns:** Error if deletion fails

**Example:**
```go
if err := storage.Delete([]byte("block:123")); err != nil {
    log.Printf("Failed to delete block: %v", err)
}
```

#### `(s *Storage) Close() error`
Closes the storage and releases resources.

**Returns:** Error if close fails

**Example:**
```go
if err := storage.Close(); err != nil {
    log.Printf("Failed to close storage: %v", err)
}
```

## Error Handling

All functions return errors that should be checked and handled appropriately. Common error patterns:

```go
// Check for errors
if err != nil {
    log.Printf("Operation failed: %v", err)
    return err
}

// Handle specific error types
if errors.Is(err, ErrBlockNotFound) {
    log.Println("Block not found")
} else if errors.Is(err, ErrInvalidSignature) {
    log.Println("Invalid signature")
}
```

## Best Practices

1. **Always check errors** returned by functions
2. **Use proper logging** for debugging and monitoring
3. **Validate inputs** before processing
4. **Handle edge cases** gracefully
5. **Use context** for cancellation and timeouts
6. **Implement proper cleanup** in defer statements
7. **Use interfaces** for testing and flexibility

## Examples

### Complete Transaction Flow

```go
func main() {
    // Create wallet
    wallet, err := wallet.NewWallet()
    if err != nil {
        log.Fatalf("Failed to create wallet: %v", err)
    }

    // Generate key pair
    privKey, pubKey, err := wallet.GenerateKeyPair()
    if err != nil {
        log.Fatalf("Failed to generate keys: %v", err)
    }

    // Get address
    address, err := wallet.GetAddress(pubKey)
    if err != nil {
        log.Fatalf("Failed to get address: %v", err)
    }

    // Create transaction
    tx, err := wallet.CreateTransaction(address, "destination", 1000000, 1000)
    if err != nil {
        log.Fatalf("Failed to create transaction: %v", err)
    }

    // Sign transaction
    if err := wallet.SignTransaction(tx, privKey); err != nil {
        log.Fatalf("Failed to sign transaction: %v", err)
    }

    // Validate transaction
    if err := tx.Validate(); err != nil {
        log.Fatalf("Transaction validation failed: %v", err)
    }

    log.Printf("Transaction created successfully: %x", tx.ID)
}
```

### Blockchain Operations

```go
func main() {
    // Create blockchain
    chain, err := chain.NewChain("./data")
    if err != nil {
        log.Fatalf("Failed to create chain: %v", err)
    }

    // Create miner
    miner := miner.NewMiner(chain)

    // Mine block
    block, err := miner.MineBlock([]*block.Transaction{})
    if err != nil {
        log.Fatalf("Mining failed: %v", err)
    }

    // Add block to chain
    if err := chain.AddBlock(block); err != nil {
        log.Fatalf("Failed to add block: %v", err)
    }

    log.Printf("Block added successfully: %x", block.Hash)
}
```

---

For more detailed information about specific functions and types, refer to the GoDoc comments in the source code. 