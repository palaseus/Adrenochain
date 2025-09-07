import axios, { AxiosInstance } from 'axios';
import { Wallet } from './wallet';
import { Transaction } from './transaction';
import { DeFi } from './defi';
import { Explorer } from './explorer';
import { types } from './types';

/**
 * AdrenochainClient - Main client for interacting with Adrenochain blockchain
 */
export class AdrenochainClient {
  private api: AxiosInstance;
  private wallet?: Wallet;
  private defi?: DeFi;
  private explorer?: Explorer;

  constructor(
    private baseUrl: string = 'http://localhost:8080',
    private apiKey?: string
  ) {
    this.api = axios.create({
      baseURL: baseUrl,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        ...(apiKey && { 'Authorization': `Bearer ${apiKey}` })
      }
    });

    // Initialize modules
    this.wallet = new Wallet(this.api);
    this.defi = new DeFi(this.api);
    this.explorer = new Explorer(this.api);
  }

  /**
   * Get wallet instance
   */
  getWallet(): Wallet {
    if (!this.wallet) {
      this.wallet = new Wallet(this.api);
    }
    return this.wallet;
  }

  /**
   * Get DeFi instance
   */
  getDeFi(): DeFi {
    if (!this.defi) {
      this.defi = new DeFi(this.api);
    }
    return this.defi;
  }

  /**
   * Get explorer instance
   */
  getExplorer(): Explorer {
    if (!this.explorer) {
      this.explorer = new Explorer(this.api);
    }
    return this.explorer;
  }

  /**
   * Get blockchain information
   */
  async getBlockchainInfo(): Promise<types.BlockchainInfo> {
    const response = await this.api.get('/api/v1/blockchain/info');
    return response.data;
  }

  /**
   * Get network status
   */
  async getNetworkStatus(): Promise<types.NetworkStatus> {
    const response = await this.api.get('/api/v1/network/status');
    return response.data;
  }

  /**
   * Get health status
   */
  async getHealth(): Promise<types.HealthStatus> {
    const response = await this.api.get('/health');
    return response.data;
  }

  /**
   * Get metrics
   */
  async getMetrics(): Promise<types.Metrics> {
    const response = await this.api.get('/metrics');
    return response.data;
  }

  /**
   * Create a new transaction
   */
  createTransaction(): Transaction {
    return new Transaction(this.api);
  }

  /**
   * Set API key for authentication
   */
  setApiKey(apiKey: string): void {
    this.apiKey = apiKey;
    this.api.defaults.headers['Authorization'] = `Bearer ${apiKey}`;
  }

  /**
   * Get current base URL
   */
  getBaseUrl(): string {
    return this.baseUrl;
  }

  /**
   * Set new base URL
   */
  setBaseUrl(baseUrl: string): void {
    this.baseUrl = baseUrl;
    this.api.defaults.baseURL = baseUrl;
  }
}
