import { AxiosInstance } from 'axios';
import { Block, Transaction, ExplorerConfig } from './types';

/**
 * Explorer class for blockchain data exploration
 */
export class Explorer {
  constructor(private api: AxiosInstance) {}

  // ===== BLOCKS =====

  /**
   * Get block by hash or number
   */
  async getBlock(blockId: string | number): Promise<Block> {
    const response = await this.api.get(`/api/v1/explorer/block/${blockId}`);
    return response.data;
  }

  /**
   * Get latest blocks
   */
  async getLatestBlocks(limit: number = 10): Promise<Block[]> {
    const response = await this.api.get('/api/v1/explorer/blocks/latest', {
      params: { limit }
    });
    return response.data.blocks;
  }

  /**
   * Get blocks by range
   */
  async getBlocksByRange(startBlock: number, endBlock: number): Promise<Block[]> {
    const response = await this.api.get('/api/v1/explorer/blocks/range', {
      params: { startBlock, endBlock }
    });
    return response.data.blocks;
  }

  /**
   * Get block count
   */
  async getBlockCount(): Promise<number> {
    const response = await this.api.get('/api/v1/explorer/blocks/count');
    return response.data.count;
  }

  /**
   * Get block statistics
   */
  async getBlockStats(): Promise<{
    totalBlocks: number;
    averageBlockTime: number;
    averageBlockSize: number;
    totalTransactions: number;
    averageTransactionsPerBlock: number;
  }> {
    const response = await this.api.get('/api/v1/explorer/blocks/stats');
    return response.data;
  }

  // ===== TRANSACTIONS =====

  /**
   * Get transaction by hash
   */
  async getTransaction(txHash: string): Promise<Transaction> {
    const response = await this.api.get(`/api/v1/explorer/transaction/${txHash}`);
    return response.data;
  }

  /**
   * Get latest transactions
   */
  async getLatestTransactions(limit: number = 10): Promise<Transaction[]> {
    const response = await this.api.get('/api/v1/explorer/transactions/latest', {
      params: { limit }
    });
    return response.data.transactions;
  }

  /**
   * Get transactions by address
   */
  async getTransactionsByAddress(address: string, config: ExplorerConfig = {}): Promise<{
    transactions: Transaction[];
    total: number;
    page: number;
    limit: number;
  }> {
    const response = await this.api.get(`/api/v1/explorer/transactions/address/${address}`, {
      params: config
    });
    return response.data;
  }

  /**
   * Get transactions by block
   */
  async getTransactionsByBlock(blockHash: string): Promise<Transaction[]> {
    const response = await this.api.get(`/api/v1/explorer/transactions/block/${blockHash}`);
    return response.data.transactions;
  }

  /**
   * Search transactions
   */
  async searchTransactions(query: string, config: ExplorerConfig = {}): Promise<{
    transactions: Transaction[];
    total: number;
  }> {
    const response = await this.api.get('/api/v1/explorer/transactions/search', {
      params: { query, ...config }
    });
    return response.data;
  }

  // ===== ACCOUNTS =====

  /**
   * Get account information
   */
  async getAccount(address: string): Promise<{
    address: string;
    balance: string;
    nonce: number;
    transactionCount: number;
    firstSeen: string;
    lastSeen: string;
  }> {
    const response = await this.api.get(`/api/v1/explorer/account/${address}`);
    return response.data;
  }

  /**
   * Get account balance history
   */
  async getAccountBalanceHistory(address: string, period: 'day' | 'week' | 'month' | 'year'): Promise<{
    timestamps: string[];
    balances: string[];
  }> {
    const response = await this.api.get(`/api/v1/explorer/account/${address}/balance-history`, {
      params: { period }
    });
    return response.data;
  }

  /**
   * Get top accounts by balance
   */
  async getTopAccounts(limit: number = 100): Promise<{
    accounts: Array<{
      address: string;
      balance: string;
      percentage: number;
    }>;
  }> {
    const response = await this.api.get('/api/v1/explorer/accounts/top', {
      params: { limit }
    });
    return response.data;
  }

  // ===== NETWORK STATS =====

  /**
   * Get network statistics
   */
  async getNetworkStats(): Promise<{
    totalBlocks: number;
    totalTransactions: number;
    totalAddresses: number;
    totalSupply: string;
    circulatingSupply: string;
    networkHashRate: string;
    difficulty: number;
    averageBlockTime: number;
    averageTransactionFee: string;
    averageTransactionSize: number;
  }> {
    const response = await this.api.get('/api/v1/explorer/network/stats');
    return response.data;
  }

  /**
   * Get network activity
   */
  async getNetworkActivity(period: 'hour' | 'day' | 'week' | 'month'): Promise<{
    timestamps: string[];
    blocks: number[];
    transactions: number[];
    addresses: number[];
  }> {
    const response = await this.api.get('/api/v1/explorer/network/activity', {
      params: { period }
    });
    return response.data;
  }

  /**
   * Get network health
   */
  async getNetworkHealth(): Promise<{
    status: 'healthy' | 'degraded' | 'unhealthy';
    blockTime: number;
    transactionThroughput: number;
    networkLatency: number;
    peerCount: number;
    syncStatus: string;
  }> {
    const response = await this.api.get('/api/v1/explorer/network/health');
    return response.data;
  }

  // ===== SEARCH =====

  /**
   * Search the blockchain
   */
  async search(query: string): Promise<{
    blocks: Block[];
    transactions: Transaction[];
    accounts: Array<{
      address: string;
      balance: string;
    }>;
  }> {
    const response = await this.api.get('/api/v1/explorer/search', {
      params: { query }
    });
    return response.data;
  }

  /**
   * Get search suggestions
   */
  async getSearchSuggestions(query: string): Promise<string[]> {
    const response = await this.api.get('/api/v1/explorer/search/suggestions', {
      params: { query }
    });
    return response.data.suggestions;
  }

  // ===== CHARTS AND ANALYTICS =====

  /**
   * Get price chart data
   */
  async getPriceChart(period: 'hour' | 'day' | 'week' | 'month' | 'year'): Promise<{
    timestamps: string[];
    prices: number[];
    volumes: number[];
  }> {
    const response = await this.api.get('/api/v1/explorer/charts/price', {
      params: { period }
    });
    return response.data;
  }

  /**
   * Get transaction volume chart
   */
  async getTransactionVolumeChart(period: 'hour' | 'day' | 'week' | 'month'): Promise<{
    timestamps: string[];
    volumes: number[];
    counts: number[];
  }> {
    const response = await this.api.get('/api/v1/explorer/charts/transaction-volume', {
      params: { period }
    });
    return response.data;
  }

  /**
   * Get network hash rate chart
   */
  async getHashRateChart(period: 'hour' | 'day' | 'week' | 'month'): Promise<{
    timestamps: string[];
    hashRates: number[];
  }> {
    const response = await this.api.get('/api/v1/explorer/charts/hash-rate', {
      params: { period }
    });
    return response.data;
  }

  // ===== TOKENS =====

  /**
   * Get token information
   */
  async getToken(tokenAddress: string): Promise<{
    address: string;
    name: string;
    symbol: string;
    decimals: number;
    totalSupply: string;
    circulatingSupply: string;
    holders: number;
    transfers: number;
  }> {
    const response = await this.api.get(`/api/v1/explorer/token/${tokenAddress}`);
    return response.data;
  }

  /**
   * Get token holders
   */
  async getTokenHolders(tokenAddress: string, config: ExplorerConfig = {}): Promise<{
    holders: Array<{
      address: string;
      balance: string;
      percentage: number;
    }>;
    total: number;
  }> {
    const response = await this.api.get(`/api/v1/explorer/token/${tokenAddress}/holders`, {
      params: config
    });
    return response.data;
  }

  /**
   * Get token transfers
   */
  async getTokenTransfers(tokenAddress: string, config: ExplorerConfig = {}): Promise<{
    transfers: Array<{
      from: string;
      to: string;
      amount: string;
      txHash: string;
      timestamp: string;
    }>;
    total: number;
  }> {
    const response = await this.api.get(`/api/v1/explorer/token/${tokenAddress}/transfers`, {
      params: config
    });
    return response.data;
  }

  // ===== CONTRACTS =====

  /**
   * Get contract information
   */
  async getContract(contractAddress: string): Promise<{
    address: string;
    name: string;
    type: string;
    verified: boolean;
    sourceCode?: string;
    abi?: any;
    creationTx: string;
    creator: string;
  }> {
    const response = await this.api.get(`/api/v1/explorer/contract/${contractAddress}`);
    return response.data;
  }

  /**
   * Get contract transactions
   */
  async getContractTransactions(contractAddress: string, config: ExplorerConfig = {}): Promise<{
    transactions: Transaction[];
    total: number;
  }> {
    const response = await this.api.get(`/api/v1/explorer/contract/${contractAddress}/transactions`, {
      params: config
    });
    return response.data;
  }

  /**
   * Verify contract
   */
  async verifyContract(contractAddress: string, sourceCode: string, compilerVersion: string): Promise<{
    verified: boolean;
    message: string;
  }> {
    const response = await this.api.post('/api/v1/explorer/contract/verify', {
      contractAddress,
      sourceCode,
      compilerVersion
    });
    return response.data;
  }
}
