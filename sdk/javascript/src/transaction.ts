import { AxiosInstance } from 'axios';
import { TransactionConfig, Transaction as TransactionType } from './types';

/**
 * Transaction class for creating and managing blockchain transactions
 */
export class Transaction {
  constructor(private api: AxiosInstance) {}

  /**
   * Create a new transaction
   */
  async createTransaction(config: TransactionConfig): Promise<TransactionType> {
    const response = await this.api.post('/api/v1/transaction/create', config);
    return response.data;
  }

  /**
   * Send a transaction
   */
  async sendTransaction(transaction: TransactionType): Promise<string> {
    const response = await this.api.post('/api/v1/transaction/send', transaction);
    return response.data.txHash;
  }

  /**
   * Get transaction by hash
   */
  async getTransaction(txHash: string): Promise<TransactionType> {
    const response = await this.api.get(`/api/v1/transaction/${txHash}`);
    return response.data;
  }

  /**
   * Get transaction status
   */
  async getTransactionStatus(txHash: string): Promise<'pending' | 'confirmed' | 'failed'> {
    const response = await this.api.get(`/api/v1/transaction/${txHash}/status`);
    return response.data.status;
  }

  /**
   * Get transaction receipt
   */
  async getTransactionReceipt(txHash: string): Promise<any> {
    const response = await this.api.get(`/api/v1/transaction/${txHash}/receipt`);
    return response.data;
  }

  /**
   * Estimate transaction fee
   */
  async estimateFee(config: TransactionConfig): Promise<string> {
    const response = await this.api.post('/api/v1/transaction/estimate-fee', config);
    return response.data.fee;
  }

  /**
   * Get pending transactions
   */
  async getPendingTransactions(): Promise<TransactionType[]> {
    const response = await this.api.get('/api/v1/transaction/pending');
    return response.data.transactions;
  }

  /**
   * Get transactions by address
   */
  async getTransactionsByAddress(address: string, page: number = 1, limit: number = 50): Promise<{
    transactions: TransactionType[];
    total: number;
    page: number;
    limit: number;
  }> {
    const response = await this.api.get(`/api/v1/transaction/address/${address}`, {
      params: { page, limit }
    });
    return response.data;
  }

  /**
   * Get transactions by block
   */
  async getTransactionsByBlock(blockHash: string): Promise<TransactionType[]> {
    const response = await this.api.get(`/api/v1/transaction/block/${blockHash}`);
    return response.data.transactions;
  }

  /**
   * Sign a transaction
   */
  async signTransaction(transaction: TransactionType, privateKey: string): Promise<TransactionType> {
    const response = await this.api.post('/api/v1/transaction/sign', {
      transaction,
      privateKey
    });
    return response.data;
  }

  /**
   * Validate a transaction
   */
  async validateTransaction(transaction: TransactionType): Promise<{
    valid: boolean;
    errors: string[];
  }> {
    const response = await this.api.post('/api/v1/transaction/validate', transaction);
    return response.data;
  }

  /**
   * Broadcast a transaction to the network
   */
  async broadcastTransaction(transaction: TransactionType): Promise<string> {
    const response = await this.api.post('/api/v1/transaction/broadcast', transaction);
    return response.data.txHash;
  }

  /**
   * Cancel a pending transaction
   */
  async cancelTransaction(txHash: string, privateKey: string): Promise<string> {
    const response = await this.api.post('/api/v1/transaction/cancel', {
      txHash,
      privateKey
    });
    return response.data.txHash;
  }

  /**
   * Speed up a pending transaction
   */
  async speedUpTransaction(txHash: string, newFee: string, privateKey: string): Promise<string> {
    const response = await this.api.post('/api/v1/transaction/speed-up', {
      txHash,
      newFee,
      privateKey
    });
    return response.data.txHash;
  }

  /**
   * Get transaction history
   */
  async getTransactionHistory(address: string, options: {
    from?: string;
    to?: string;
    limit?: number;
    offset?: number;
  } = {}): Promise<{
    transactions: TransactionType[];
    total: number;
    hasMore: boolean;
  }> {
    const response = await this.api.get(`/api/v1/transaction/history/${address}`, {
      params: options
    });
    return response.data;
  }

  /**
   * Create a batch transaction
   */
  async createBatchTransaction(transactions: TransactionConfig[]): Promise<TransactionType[]> {
    const response = await this.api.post('/api/v1/transaction/batch', {
      transactions
    });
    return response.data;
  }

  /**
   * Get transaction analytics
   */
  async getTransactionAnalytics(address: string, period: 'day' | 'week' | 'month' | 'year'): Promise<{
    totalTransactions: number;
    totalVolume: string;
    averageFee: string;
    gasUsed: number;
    successRate: number;
  }> {
    const response = await this.api.get(`/api/v1/transaction/analytics/${address}`, {
      params: { period }
    });
    return response.data;
  }
}
