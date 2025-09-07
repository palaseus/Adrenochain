# Adrenochain JavaScript/TypeScript SDK

A comprehensive SDK for interacting with the Adrenochain blockchain network, providing easy-to-use interfaces for wallet management, transaction creation, DeFi operations, and blockchain exploration.

## 🚀 Quick Start

### Installation

```bash
npm install adrenochain-sdk
```

### Basic Usage

```typescript
import { AdrenochainClient } from 'adrenochain-sdk';

// Create a client instance
const client = new AdrenochainClient('http://localhost:8080');

// Get blockchain information
const info = await client.getBlockchainInfo();
console.log('Blockchain height:', info.height);

// Get wallet instance
const wallet = client.getWallet();

// Create a new wallet
const account = await wallet.createWallet({
  passphrase: 'my-secure-password'
});

console.log('New wallet address:', account.address);
```

## 📚 API Reference

### AdrenochainClient

The main client class for interacting with the Adrenochain network.

```typescript
const client = new AdrenochainClient(baseUrl?, apiKey?);
```

#### Methods

- `getBlockchainInfo()` - Get blockchain information
- `getNetworkStatus()` - Get network status
- `getHealth()` - Get health status
- `getMetrics()` - Get system metrics
- `getWallet()` - Get wallet instance
- `getDeFi()` - Get DeFi instance
- `getExplorer()` - Get explorer instance
- `createTransaction()` - Create a new transaction

### Wallet

Manage accounts and transactions.

```typescript
const wallet = client.getWallet();

// Create wallet
await wallet.createWallet({
  passphrase: 'password',
  derivationPath: "m/44'/60'/0'/0/0"
});

// Get balance
const balance = await wallet.getBalance(address);

// Sign message
const signature = await wallet.signMessage(address, message, passphrase);
```

### Transaction

Create and manage blockchain transactions.

```typescript
const tx = client.createTransaction();

// Create transaction
const transaction = await tx.createTransaction({
  to: '0x123...',
  value: '1000000000000000000',
  fee: '1000000000000000'
});

// Send transaction
const txHash = await tx.sendTransaction(transaction);

// Get transaction status
const status = await tx.getTransactionStatus(txHash);
```

### DeFi

Interact with decentralized finance protocols.

```typescript
const defi = client.getDeFi();

// Get pools
const pools = await defi.getPools();

// Add liquidity
const position = await defi.addLiquidity({
  poolId: 'pool-123',
  amountA: '1000000000000000000',
  amountB: '2000000000000000000'
});

// Swap tokens
const txHash = await defi.swap(
  'tokenA',
  'tokenB',
  '1000000000000000000',
  '1900000000000000000'
);
```

### Explorer

Explore blockchain data.

```typescript
const explorer = client.getExplorer();

// Get block
const block = await explorer.getBlock(12345);

// Get transaction
const tx = await explorer.getTransaction('0xabc...');

// Get account info
const account = await explorer.getAccount('0x123...');

// Search
const results = await explorer.search('0x123...');
```

## 🔧 Configuration

### Environment Variables

```bash
ADRENOCHAIN_API_URL=http://localhost:8080
ADRENOCHAIN_API_KEY=your-api-key
```

### Custom Configuration

```typescript
const client = new AdrenochainClient('http://localhost:8080', 'api-key');

// Set custom headers
client.setApiKey('new-api-key');

// Change base URL
client.setBaseUrl('https://api.adrenochain.com');
```

## 📦 Examples

### Complete Wallet Management

```typescript
import { AdrenochainClient } from 'adrenochain-sdk';

async function walletExample() {
  const client = new AdrenochainClient('http://localhost:8080');
  const wallet = client.getWallet();

  // Create wallet
  const account = await wallet.createWallet({
    passphrase: 'secure-password'
  });

  console.log('Wallet created:', account.address);

  // Get balance
  const balance = await wallet.getBalance(account.address);
  console.log('Balance:', balance);

  // Generate new address
  const newAddress = await wallet.generateAddress('secure-password');
  console.log('New address:', newAddress);
}
```

### DeFi Operations

```typescript
async function defiExample() {
  const client = new AdrenochainClient('http://localhost:8080');
  const defi = client.getDeFi();

  // Get available pools
  const pools = await defi.getPools();
  console.log('Available pools:', pools);

  // Add liquidity to first pool
  if (pools.length > 0) {
    const position = await defi.addLiquidity({
      poolId: pools[0].id,
      amountA: '1000000000000000000',
      amountB: '2000000000000000000'
    });
    console.log('Liquidity added:', position.id);
  }
}
```

### Transaction Management

```typescript
async function transactionExample() {
  const client = new AdrenochainClient('http://localhost:8080');
  const tx = client.createTransaction();

  // Create transaction
  const transaction = await tx.createTransaction({
    to: '0x1234567890123456789012345678901234567890',
    value: '1000000000000000000', // 1 token
    fee: '1000000000000000' // 0.001 token
  });

  // Send transaction
  const txHash = await tx.sendTransaction(transaction);
  console.log('Transaction sent:', txHash);

  // Wait for confirmation
  let status = 'pending';
  while (status === 'pending') {
    await new Promise(resolve => setTimeout(resolve, 1000));
    status = await tx.getTransactionStatus(txHash);
    console.log('Status:', status);
  }
}
```

## 🧪 Testing

```bash
# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run linting
npm run lint

# Format code
npm run format
```

## 📖 TypeScript Support

This SDK is written in TypeScript and provides full type definitions.

```typescript
import { 
  AdrenochainClient, 
  Wallet, 
  Transaction, 
  DeFi, 
  Explorer,
  types 
} from 'adrenochain-sdk';

// Use types for better development experience
const config: types.TransactionConfig = {
  to: '0x123...',
  value: '1000000000000000000',
  fee: '1000000000000000'
};
```

## 🔒 Security

- All private keys are handled securely
- API keys are transmitted over HTTPS
- Wallet passphrases are never stored
- All transactions are cryptographically signed

## 📝 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📞 Support

- Documentation: [docs.adrenochain.com](https://docs.adrenochain.com)
- Discord: [discord.gg/adrenochain](https://discord.gg/adrenochain)
- Email: support@adrenochain.com
