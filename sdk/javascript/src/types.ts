/**
 * Type definitions for Adrenochain SDK
 */

export interface BlockchainInfo {
  chainId: string;
  height: number;
  bestBlockHash: string;
  totalTransactions: number;
  totalSupply: string;
  blockTime: number;
  difficulty: number;
  networkHashRate: string;
}

export interface NetworkStatus {
  connectedPeers: number;
  totalPeers: number;
  networkId: string;
  protocolVersion: string;
  isSyncing: boolean;
  syncProgress: number;
}

export interface HealthStatus {
  status: 'healthy' | 'unhealthy' | 'degraded';
  timestamp: string;
  uptime: number;
  version: string;
  services: {
    blockchain: boolean;
    network: boolean;
    storage: boolean;
    api: boolean;
  };
}

export interface Metrics {
  blocks: {
    total: number;
    mined: number;
    pending: number;
  };
  transactions: {
    total: number;
    pending: number;
    confirmed: number;
  };
  network: {
    peers: number;
    connections: number;
    bytesReceived: number;
    bytesSent: number;
  };
  performance: {
    blockProcessingTime: number;
    transactionProcessingTime: number;
    memoryUsage: number;
    cpuUsage: number;
  };
}

export interface Account {
  address: string;
  balance: string;
  nonce: number;
  publicKey?: string;
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  value: string;
  fee: string;
  nonce: number;
  timestamp: string;
  status: 'pending' | 'confirmed' | 'failed';
  blockNumber?: number;
  blockHash?: string;
}

export interface Block {
  hash: string;
  number: number;
  parentHash: string;
  timestamp: string;
  difficulty: number;
  nonce: number;
  transactions: Transaction[];
  miner: string;
  size: number;
}

export interface DeFiPool {
  id: string;
  tokenA: string;
  tokenB: string;
  reserveA: string;
  reserveB: string;
  totalSupply: string;
  fee: number;
  apr: number;
}

export interface DeFiPosition {
  id: string;
  poolId: string;
  owner: string;
  liquidity: string;
  tokenA: string;
  tokenB: string;
  value: string;
  fees: string;
}

export interface GovernanceProposal {
  id: number;
  title: string;
  description: string;
  proposer: string;
  startTime: string;
  endTime: string;
  status: 'active' | 'passed' | 'rejected' | 'executed';
  forVotes: string;
  againstVotes: string;
  abstainVotes: string;
}

export interface WalletConfig {
  privateKey?: string;
  mnemonic?: string;
  passphrase?: string;
  derivationPath?: string;
}

export interface TransactionConfig {
  to: string;
  value: string;
  fee?: string;
  nonce?: number;
  data?: string;
}

export interface DeFiConfig {
  poolId: string;
  amountA: string;
  amountB: string;
  slippageTolerance?: number;
  deadline?: number;
}

export interface ExplorerConfig {
  page?: number;
  limit?: number;
  sort?: 'asc' | 'desc';
  filter?: Record<string, any>;
}

export namespace types {
  export type {
    BlockchainInfo,
    NetworkStatus,
    HealthStatus,
    Metrics,
    Account,
    Transaction,
    Block,
    DeFiPool,
    DeFiPosition,
    GovernanceProposal,
    WalletConfig,
    TransactionConfig,
    DeFiConfig,
    ExplorerConfig
  };
}
