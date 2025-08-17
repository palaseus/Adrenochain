# adrenochain API Documentation

## Overview

adrenochain provides a comprehensive API for blockchain operations, DeFi protocols, cross-chain bridging, and governance. This document covers all available endpoints and their usage.

## Table of Contents

1. [Exchange Layer API](#exchange-layer-api)
2. [Bridge Infrastructure API](#bridge-infrastructure-api)
3. [Governance & DAO API](#governance--dao-api)
4. [DeFi Protocols API](#defi-protocols-api)
5. [Core Blockchain API](#core-blockchain-api)
6. [WebSocket APIs](#websocket-apis)

## Exchange Layer API

### Trading API

The trading API provides endpoints for order management, market data, and trading operations.

#### Base URL
```
http://localhost:8080/api/v1
```

#### Endpoints

##### Create Order
```http
POST /orders
```

**Request Body:**
```json
{
  "trading_pair": "BTC/USDT",
  "side": "buy",
  "type": "limit",
  "quantity": "1000000",
  "price": "50000",
  "user_id": "user123",
  "time_in_force": "GTC"
}
```

**Response:**
```json
{
  "order_id": "order_12345",
  "status": "pending",
  "message": "Order created successfully"
}
```

##### Get Order
```http
GET /orders/{order_id}
```

**Response:**
```json
{
  "order": {
    "id": "order_12345",
    "trading_pair": "BTC/USDT",
    "side": "buy",
    "type": "limit",
    "quantity": "1000000",
    "price": "50000",
    "user_id": "user123",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

##### Cancel Order
```http
DELETE /orders/{order_id}
```

**Response:**
```json
{
  "status": "cancelled",
  "message": "Order cancelled successfully"
}
```

##### Get Order Book
```http
GET /orderbook/{trading_pair}?depth=10
```

**Response:**
```json
{
  "trading_pair": "BTC/USDT",
  "bids": [
    {
      "price": "50000",
      "quantity": "1000000",
      "total": "1000000"
    }
  ],
  "asks": [
    {
      "price": "50100",
      "quantity": "500000",
      "total": "500000"
    }
  ],
  "last_updated": "2024-01-01T00:00:00Z"
}
```

##### Get Trading Pairs
```http
GET /trading-pairs
```

**Response:**
```json
{
  "pairs": [
    {
      "base_asset": "BTC",
      "quote_asset": "USDT",
      "min_quantity": "1000",
      "max_quantity": "1000000",
      "min_price": "100",
      "max_price": "100000",
      "tick_size": "1",
      "step_size": "1",
      "maker_fee": "100",
      "taker_fee": "200",
      "status": "active"
    }
  ]
}
```

##### Get Market Data
```http
GET /market-data/{trading_pair}
```

**Response:**
```json
{
  "trading_pair": "BTC/USDT",
  "last_price": "50000",
  "price_change_24h": "500",
  "volume_24h": "1000000000",
  "high_24h": "51000",
  "low_24h": "49000",
  "bid": "49900",
  "ask": "50100"
}
```

### WebSocket Market Data

#### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/market-data');
```

#### Subscribe to Order Book Updates
```json
{
  "type": "subscribe",
  "channel": "orderbook",
  "trading_pair": "BTC/USDT"
}
```

#### Subscribe to Trade Updates
```json
{
  "type": "subscribe",
  "channel": "trades",
  "trading_pair": "BTC/USDT"
}
```

#### Subscribe to Market Data Updates
```json
{
  "type": "subscribe",
  "channel": "market_data",
  "trading_pair": "BTC/USDT"
}
```

## Advanced Derivatives & Risk Management API 🚀

### Options Trading

#### Create Options Contract
```http
POST /derivatives/options
```

**Request Body:**
```json
{
  "type": "european",
  "underlying": "BTC/USDT",
  "strike_price": "50000",
  "expiration": "2024-12-31T23:59:59Z",
  "option_type": "call",
  "quantity": "1000000",
  "user_id": "user123"
}
```

**Response:**
```json
{
  "contract_id": "opt_12345",
  "status": "active",
  "premium": "2500",
  "delta": "0.65",
  "gamma": "0.02",
  "theta": "-15.5",
  "vega": "120.3",
  "message": "Options contract created successfully"
}
```

#### Get Options Greeks
```http
GET /derivatives/options/{contract_id}/greeks
```

**Response:**
```json
{
  "contract_id": "opt_12345",
  "delta": "0.65",
  "gamma": "0.02",
  "theta": "-15.5",
  "vega": "120.3",
  "rho": "8.2",
  "calculated_at": "2024-01-01T00:00:00Z"
}
```

#### Exercise Options Contract
```http
POST /derivatives/options/{contract_id}/exercise
```

**Request Body:**
```json
{
  "user_id": "user123",
  "quantity": "1000000"
}
```

### Futures Trading

#### Create Futures Contract
```http
POST /derivatives/futures
```

**Request Body:**
```json
{
  "type": "perpetual",
  "underlying": "BTC/USDT",
  "leverage": "10",
  "quantity": "1000000",
  "side": "long",
  "user_id": "user123"
}
```

**Response:**
```json
{
  "contract_id": "fut_12345",
  "status": "active",
  "entry_price": "50000",
  "funding_rate": "0.0001",
  "liquidation_price": "45000",
  "margin_required": "5000",
  "message": "Futures contract created successfully"
}
```

#### Get Funding Rate
```http
GET /derivatives/futures/{contract_id}/funding-rate
```

**Response:**
```json
{
  "contract_id": "fut_12345",
  "current_rate": "0.0001",
  "next_rate": "0.0002",
  "next_funding_time": "2024-01-01T08:00:00Z",
  "last_updated": "2024-01-01T00:00:00Z"
}
```

### Risk Management

#### Calculate VaR
```http
POST /risk/var
```

**Request Body:**
```json
{
  "portfolio": [
    {
      "asset": "BTC",
      "quantity": "1.5",
      "current_price": "50000"
    },
    {
      "asset": "ETH",
      "quantity": "25.0",
      "current_price": "3000"
    }
  ],
  "confidence_level": "0.95",
  "time_horizon": "1",
  "method": "historical_simulation"
}
```

**Response:**
```json
{
  "var_95": "12500",
  "var_99": "18500",
  "expected_shortfall": "15000",
  "confidence_interval": [12000, 13000],
  "calculated_at": "2024-01-01T00:00:00Z"
}
```

#### Stress Testing
```http
POST /risk/stress-test
```

**Request Body:**
```json
{
  "portfolio_id": "port_12345",
  "scenarios": [
    {
      "name": "Market Crash",
      "btc_price_change": "-30%",
      "eth_price_change": "-25%",
      "volatility_change": "+50%"
    },
    {
      "name": "Interest Rate Shock",
      "rate_change": "+2%",
      "correlation_change": "+0.2"
    }
  ]
}
```

**Response:**
```json
{
  "portfolio_id": "port_12345",
  "scenarios": [
    {
      "name": "Market Crash",
      "portfolio_value": "67500",
      "loss": "22500",
      "var_impact": "18500"
    },
    {
      "name": "Interest Rate Shock",
      "portfolio_value": "85000",
      "loss": "5000",
      "var_impact": "3500"
    }
  ]
}
```

### Insurance Protocols

#### Create Coverage Pool
```http
POST /insurance/coverage-pools
```

**Request Body:**
```json
{
  "name": "DeFi Protocol Insurance",
  "coverage_type": "smart_contract_risk",
  "max_coverage": "1000000",
  "premium_rate": "0.05",
  "user_id": "user123"
}
```

**Response:**
```json
{
  "pool_id": "ins_12345",
  "status": "active",
  "total_coverage": "1000000",
  "available_coverage": "1000000",
  "total_premiums": "50000",
  "message": "Coverage pool created successfully"
}
```

#### Submit Insurance Claim
```http
POST /insurance/claims
```

**Request Body:**
```json
{
  "pool_id": "ins_12345",
  "user_id": "user123",
  "claim_amount": "50000",
  "description": "Smart contract exploit resulting in fund loss",
  "evidence": "transaction_hash_12345"
}
```

### Liquidation Systems

#### Get Liquidation Status
```http
GET /liquidation/status/{position_id}
```

**Response:**
```json
{
  "position_id": "pos_12345",
  "status": "at_risk",
  "collateral_ratio": "1.15",
  "liquidation_threshold": "1.1",
  "liquidation_price": "45000",
  "time_to_liquidation": "2h",
  "recommended_action": "add_collateral"
}
```

#### Liquidate Position
```http
POST /liquidation/execute
```

**Request Body:**
```json
{
  "position_id": "pos_12345",
  "liquidator_id": "liquidator_123",
  "collateral_assets": ["BTC", "ETH"]
}
```

### Algorithmic Trading

#### Create Trading Strategy
```http
POST /trading/strategies
```

**Request Body:**
```json
{
  "name": "Mean Reversion BTC",
  "type": "mean_reversion",
  "parameters": {
    "lookback_period": "20",
    "entry_threshold": "2.0",
    "exit_threshold": "0.5",
    "position_size": "0.1"
  },
  "trading_pair": "BTC/USDT",
  "user_id": "user123"
}
```

**Response:**
```json
{
  "strategy_id": "strat_12345",
  "status": "active",
  "total_trades": "0",
  "pnl": "0",
  "sharpe_ratio": "0",
  "message": "Trading strategy created successfully"
}
```

#### Get Strategy Performance
```http
GET /trading/strategies/{strategy_id}/performance
```

**Response:**
```json
{
  "strategy_id": "strat_12345",
  "total_trades": "45",
  "winning_trades": "28",
  "losing_trades": "17",
  "total_pnl": "12500",
  "sharpe_ratio": "1.85",
  "max_drawdown": "8500",
  "win_rate": "0.62"
}
```

### Market Making

#### Create Market Making Strategy
```http
POST /trading/market-making
```

**Request Body:**
```json
{
  "trading_pair": "BTC/USDT",
  "spread_target": "0.002",
  "inventory_target": "0.5",
  "max_position": "1000000",
  "user_id": "user123"
}
```

**Response:**
```json
{
  "strategy_id": "mm_12345",
  "status": "active",
  "current_spread": "0.0018",
  "current_inventory": "0.48",
  "total_volume": "5000000",
  "message": "Market making strategy created successfully"
}
```

## Bridge Infrastructure API

### Validator Management

#### Add Validator
```http
POST /bridge/validators
```

**Request Body:**
```json
{
  "id": "validator_123",
  "chain_id": "adrenochain",
  "stake_amount": "1000000000000000000",
  "public_key": "0x..."
}
```

#### Remove Validator
```http
DELETE /bridge/validators/{validator_id}
```

#### Update Validator Stake
```http
PUT /bridge/validators/{validator_id}/stake
```

**Request Body:**
```json
{
  "new_stake_amount": "2000000000000000000"
}
```

#### Get Validator List
```http
GET /bridge/validators
```

### Cross-Chain Transactions

#### Initiate Transfer
```http
POST /bridge/transfers
```

**Request Body:**
```json
{
  "source_chain": "adrenochain",
  "destination_chain": "ethereum",
  "source_address": "0x...",
  "destination_address": "0x...",
  "asset_type": "native",
  "amount": "1000000000000000000",
  "fee": "10000000000000000"
}
```

#### Get Transfer Status
```http
GET /bridge/transfers/{transfer_id}
```

#### Batch Transfer
```http
POST /bridge/transfers/batch
```

**Request Body:**
```json
{
  "transfers": [
    {
      "source_chain": "adrenochain",
      "destination_chain": "ethereum",
      "source_address": "0x...",
      "destination_address": "0x...",
      "asset_type": "native",
      "amount": "1000000000000000000"
    }
  ]
}
```

### Bridge Security

#### Check Transfer Security
```http
POST /bridge/security/check
```

**Request Body:**
```json
{
  "source_address": "0x...",
  "destination_address": "0x...",
  "amount": "1000000000000000000",
  "asset_type": "native"
}
```

#### Emergency Pause
```http
POST /bridge/security/pause
```

#### Emergency Resume
```http
POST /bridge/security/resume
```

## Governance & DAO API

### Proposal Management

#### Create Proposal
```http
POST /governance/proposals
```

**Request Body:**
```json
{
  "title": "Increase Treasury Daily Limit",
  "description": "Proposal to increase the daily treasury spending limit",
  "proposal_type": "treasury",
  "quorum_required": "1000000000000000000",
  "min_voting_power": "100000000000000000"
}
```

#### Get Proposal
```http
GET /governance/proposals/{proposal_id}
```

#### Activate Proposal
```http
POST /governance/proposals/{proposal_id}/activate
```

#### Get All Proposals
```http
GET /governance/proposals
```

### Voting

#### Cast Vote
```http
POST /governance/proposals/{proposal_id}/vote
```

**Request Body:**
```json
{
  "voter": "0x...",
  "vote_choice": "for",
  "reason": "This proposal will improve treasury efficiency"
}
```

#### Delegate Voting Power
```http
POST /governance/delegations
```

**Request Body:**
```json
{
  "delegator": "0x...",
  "delegate": "0x...",
  "amount": "1000000000000000000"
}
```

#### Get Voting Results
```http
GET /governance/proposals/{proposal_id}/results
```

### Treasury Management

#### Create Treasury Proposal
```http
POST /treasury/proposals
```

**Request Body:**
```json
{
  "title": "Fund Development Team",
  "description": "Allocate funds for development team expansion",
  "amount": "5000000000000000000",
  "asset": "ETH",
  "recipient": "0x...",
  "purpose": "Development team expansion"
}
```

#### Execute Treasury Transaction
```http
POST /treasury/transactions
```

**Request Body:**
```json
{
  "transaction_type": "transfer",
  "amount": "1000000000000000000",
  "asset": "ETH",
  "to": "0x...",
  "description": "Development funding",
  "executor": "0x..."
}
```

#### Get Treasury Balance
```http
GET /treasury/balances
```

#### Get Treasury Transactions
```http
GET /treasury/transactions
```

## DeFi Protocols API

### Lending Protocol

#### Create Lending Pool
```http
POST /defi/lending/pools
```

**Request Body:**
```json
{
  "asset": "ETH",
  "max_supply": "100000000000000000000",
  "interest_rate": "500",
  "liquidation_threshold": "8000"
}
```

#### Supply Assets
```http
POST /defi/lending/supply
```

**Request Body:**
```json
{
  "pool_id": "pool_123",
  "amount": "1000000000000000000",
  "user": "0x..."
}
```

#### Borrow Assets
```http
POST /defi/lending/borrow
```

**Request Body:**
```json
{
  "pool_id": "pool_123",
  "amount": "500000000000000000",
  "user": "0x..."
}
```

### AMM (Automated Market Maker)

#### Create Liquidity Pool
```http
POST /defi/amm/pools
```

**Request Body:**
```json
{
  "asset_a": "ETH",
  "asset_b": "USDT",
  "fee_rate": "300"
}
```

#### Add Liquidity
```http
POST /defi/amm/liquidity/add
```

**Request Body:**
```json
{
  "pool_id": "pool_123",
  "amount_a": "1000000000000000000",
  "amount_b": "50000000000000000000"
}
```

#### Swap Tokens
```http
POST /defi/amm/swap
```

**Request Body:**
```json
{
  "pool_id": "pool_123",
  "input_asset": "ETH",
  "output_asset": "USDT",
  "input_amount": "100000000000000000",
  "min_output_amount": "4900000000000000000"
}
```

## Core Blockchain API

### Block Operations

#### Get Block by Hash
```http
GET /blocks/{block_hash}
```

#### Get Block by Number
```http
GET /blocks/number/{block_number}
```

#### Get Latest Block
```http
GET /blocks/latest
```

### Transaction Operations

#### Get Transaction
```http
GET /transactions/{tx_hash}
```

#### Send Transaction
```http
POST /transactions
```

**Request Body:**
```json
{
  "from": "0x...",
  "to": "0x...",
  "value": "1000000000000000000",
  "gas_limit": "21000",
  "gas_price": "20000000000"
}
```

### Account Operations

#### Get Account Balance
```http
GET /accounts/{address}/balance
```

#### Get Account Transactions
```http
GET /accounts/{address}/transactions
```

## WebSocket APIs

### Blockchain Events

#### Subscribe to New Blocks
```json
{
  "type": "subscribe",
  "channel": "blocks"
}
```

#### Subscribe to New Transactions
```json
{
  "type": "subscribe",
  "channel": "transactions",
  "address": "0x..."
}
```

### Bridge Events

#### Subscribe to Bridge Events
```json
{
  "type": "subscribe",
  "channel": "bridge_events"
}
```

### Governance Events

#### Subscribe to Governance Events
```json
{
  "type": "subscribe",
  "channel": "governance_events"
}
```

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request parameters",
    "details": {
      "field": "amount",
      "issue": "Amount must be positive"
    }
  }
}
```

### Common Error Codes

- `INVALID_REQUEST` - Invalid request parameters
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `RATE_LIMITED` - Too many requests
- `INTERNAL_ERROR` - Internal server error

## Rate Limiting

API endpoints are rate-limited to ensure fair usage:

- **Public endpoints**: 100 requests per minute
- **Authenticated endpoints**: 1000 requests per minute
- **Admin endpoints**: 10000 requests per minute

## Authentication

Most endpoints require authentication using API keys or JWT tokens:

```http
Authorization: Bearer <jwt_token>
```

or

```http
X-API-Key: <api_key>
```

## SDKs and Libraries

### Go SDK
```go
import "github.com/adrenochain/adrenochain/pkg/sdk"

client := sdk.NewClient("http://localhost:8080")
```

### JavaScript SDK
```javascript
import { adrenochainClient } from '@adrenochain/sdk'

const client = new adrenochainClient('http://localhost:8080')
```

## Testing

### Testnet Endpoints
- **Testnet API**: `https://testnet-api.adrenochain.io`
- **Testnet WebSocket**: `wss://testnet-ws.adrenochain.io`

### Sandbox Environment
- **Sandbox API**: `https://sandbox-api.adrenochain.io`
- **Sandbox WebSocket**: `wss://sandbox-ws.adrenochain.io`

## Support

For API support and questions:
- **Documentation**: [docs.adrenochain.io](https://docs.adrenochain.io)
- **GitHub Issues**: [github.com/adrenochain/adrenochain/issues](https://github.com/adrenochain/adrenochain/issues)
- **Discord**: [discord.gg/adrenochain](https://discord.gg/adrenochain)
- **Email**: api-support@adrenochain.io 