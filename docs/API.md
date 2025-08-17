# GoChain API Documentation

## Overview

GoChain provides a comprehensive API for blockchain operations, DeFi protocols, cross-chain bridging, and governance. This document covers all available endpoints and their usage.

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
  "chain_id": "gochain",
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
  "source_chain": "gochain",
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
      "source_chain": "gochain",
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
import "github.com/gochain/gochain/pkg/sdk"

client := sdk.NewClient("http://localhost:8080")
```

### JavaScript SDK
```javascript
import { GoChainClient } from '@gochain/sdk'

const client = new GoChainClient('http://localhost:8080')
```

## Testing

### Testnet Endpoints
- **Testnet API**: `https://testnet-api.gochain.io`
- **Testnet WebSocket**: `wss://testnet-ws.gochain.io`

### Sandbox Environment
- **Sandbox API**: `https://sandbox-api.gochain.io`
- **Sandbox WebSocket**: `wss://sandbox-ws.gochain.io`

## Support

For API support and questions:
- **Documentation**: [docs.gochain.io](https://docs.gochain.io)
- **GitHub Issues**: [github.com/gochain/gochain/issues](https://github.com/gochain/gochain/issues)
- **Discord**: [discord.gg/gochain](https://discord.gg/gochain)
- **Email**: api-support@gochain.io 