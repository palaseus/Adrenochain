# Adrenochain Web Wallet

A modern, secure web-based wallet for the Adrenochain blockchain network. Built with React, TypeScript, and Tailwind CSS.

## 🚀 Features

- **Secure Wallet Management**: Create, import, and manage Adrenochain wallets
- **Transaction Management**: Send and receive ADR tokens with ease
- **DeFi Integration**: Access DeFi protocols and liquidity pools
- **Real-time Updates**: Live balance and transaction updates
- **Responsive Design**: Works seamlessly on desktop and mobile
- **Dark Mode Support**: Toggle between light and dark themes
- **Multi-language Support**: Internationalization ready

## 🛠️ Tech Stack

- **React 18** - Modern React with hooks and concurrent features
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first CSS framework
- **React Router** - Client-side routing
- **React Query** - Data fetching and caching
- **React Hook Form** - Form management
- **Lucide React** - Beautiful icons
- **Vite** - Fast build tool and dev server

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/palaseus/adrenochain.git
cd adrenochain/apps/web-wallet

# Install dependencies
npm install

# Start development server
npm run dev
```

## 🚀 Development

```bash
# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run linting
npm run lint

# Fix linting issues
npm run lint:fix

# Type checking
npm run type-check
```

## 📱 Usage

### Creating a Wallet

1. Navigate to the wallet application
2. Click "Create New Wallet"
3. Set a strong passphrase
4. Save your private key securely
5. Start using your wallet

### Importing a Wallet

1. Click "Import Wallet"
2. Enter your private key or mnemonic phrase
3. Set a passphrase if your wallet is encrypted
4. Access your existing wallet

### Sending Transactions

1. Click "Send" in the dashboard
2. Enter recipient address
3. Specify amount and fee
4. Confirm transaction
5. Wait for confirmation

### DeFi Operations

1. Navigate to "DeFi" section
2. Browse available pools
3. Add liquidity or stake tokens
4. Monitor your positions
5. Claim rewards

## 🔒 Security

- **Local Storage**: Private keys are stored locally and encrypted
- **No Server Storage**: Wallet data never leaves your device
- **Secure Connections**: All API calls use HTTPS
- **Input Validation**: All inputs are validated and sanitized
- **Error Handling**: Comprehensive error handling and user feedback

## 🌐 API Integration

The wallet integrates with the Adrenochain API for:

- Blockchain data retrieval
- Transaction broadcasting
- Balance queries
- DeFi protocol interactions
- Network status monitoring

## 📊 Features Overview

### Dashboard
- Wallet balance display
- Recent transaction history
- Portfolio overview
- Quick action buttons

### Send/Receive
- QR code generation for receiving
- Address validation
- Fee estimation
- Transaction confirmation

### DeFi
- Liquidity pool management
- Staking operations
- Yield farming
- Governance participation

### Settings
- Wallet management
- Security settings
- Network configuration
- Theme preferences

## 🎨 Customization

### Theming

The wallet supports custom themes through Tailwind CSS:

```css
/* Custom theme variables */
:root {
  --primary-color: #3b82f6;
  --secondary-color: #6b7280;
  --success-color: #10b981;
  --warning-color: #f59e0b;
  --error-color: #ef4444;
}
```

### Styling

Components use Tailwind utility classes and can be easily customized:

```tsx
// Custom button component
<button className="btn-primary">
  Send Transaction
</button>
```

## 📱 Mobile Support

The wallet is fully responsive and optimized for mobile devices:

- Touch-friendly interface
- Mobile-optimized forms
- Responsive navigation
- Mobile-specific features

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory:

```env
VITE_API_URL=http://localhost:8080
VITE_NETWORK_ID=adrenochain-mainnet
VITE_CHAIN_ID=1
```

### API Configuration

Configure the API endpoint in `src/config/api.ts`:

```typescript
export const API_CONFIG = {
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 30000,
  retries: 3
}
```

## 🧪 Testing

```bash
# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run e2e tests
npm run test:e2e
```

## 📦 Building

```bash
# Build for production
npm run build

# Build for staging
npm run build:staging

# Analyze bundle
npm run build:analyze
```

## 🚀 Deployment

### Vercel

```bash
# Deploy to Vercel
vercel --prod
```

### Netlify

```bash
# Build and deploy
npm run build
netlify deploy --prod --dir=dist
```

### Docker

```bash
# Build Docker image
docker build -t adrenochain-wallet .

# Run container
docker run -p 3000:3000 adrenochain-wallet
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🆘 Support

- Documentation: [docs.adrenochain.com](https://docs.adrenochain.com)
- Discord: [discord.gg/adrenochain](https://discord.gg/adrenochain)
- Email: support@adrenochain.com

## 🔮 Roadmap

- [ ] Hardware wallet integration
- [ ] Multi-signature support
- [ ] Advanced DeFi features
- [ ] Mobile app
- [ ] Browser extension
- [ ] Cross-chain support
