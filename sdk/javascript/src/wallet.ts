import { AxiosInstance } from 'axios';
import { WalletConfig, Account } from './types';
import * as crypto from 'crypto-js';

/**
 * Wallet class for managing Adrenochain accounts and transactions
 */
export class Wallet {
  constructor(private api: AxiosInstance) {}

  /**
   * Create a new wallet
   */
  async createWallet(config: WalletConfig): Promise<Account> {
    const response = await this.api.post('/api/v1/wallet/create', {
      mnemonic: config.mnemonic,
      passphrase: config.passphrase,
      derivationPath: config.derivationPath || "m/44'/60'/0'/0/0"
    });

    return response.data;
  }

  /**
   * Import wallet from private key
   */
  async importWallet(privateKey: string, passphrase?: string): Promise<Account> {
    const response = await this.api.post('/api/v1/wallet/import', {
      privateKey,
      passphrase
    });

    return response.data;
  }

  /**
   * Get wallet balance
   */
  async getBalance(address: string): Promise<string> {
    const response = await this.api.get(`/api/v1/wallet/balance/${address}`);
    return response.data.balance;
  }

  /**
   * Get account information
   */
  async getAccount(address: string): Promise<Account> {
    const response = await this.api.get(`/api/v1/wallet/account/${address}`);
    return response.data;
  }

  /**
   * Get account nonce
   */
  async getNonce(address: string): Promise<number> {
    const response = await this.api.get(`/api/v1/wallet/nonce/${address}`);
    return response.data.nonce;
  }

  /**
   * Sign a message with the wallet
   */
  async signMessage(address: string, message: string, passphrase?: string): Promise<string> {
    const response = await this.api.post('/api/v1/wallet/sign', {
      address,
      message,
      passphrase
    });

    return response.data.signature;
  }

  /**
   * Verify a signature
   */
  async verifySignature(address: string, message: string, signature: string): Promise<boolean> {
    const response = await this.api.post('/api/v1/wallet/verify', {
      address,
      message,
      signature
    });

    return response.data.valid;
  }

  /**
   * Generate a new address from the wallet
   */
  async generateAddress(passphrase?: string): Promise<string> {
    const response = await this.api.post('/api/v1/wallet/generate-address', {
      passphrase
    });

    return response.data.address;
  }

  /**
   * Get all addresses in the wallet
   */
  async getAddresses(): Promise<string[]> {
    const response = await this.api.get('/api/v1/wallet/addresses');
    return response.data.addresses;
  }

  /**
   * Export wallet private key
   */
  async exportPrivateKey(address: string, passphrase?: string): Promise<string> {
    const response = await this.api.post('/api/v1/wallet/export', {
      address,
      passphrase
    });

    return response.data.privateKey;
  }

  /**
   * Lock the wallet
   */
  async lockWallet(): Promise<void> {
    await this.api.post('/api/v1/wallet/lock');
  }

  /**
   * Unlock the wallet
   */
  async unlockWallet(passphrase: string): Promise<void> {
    await this.api.post('/api/v1/wallet/unlock', { passphrase });
  }

  /**
   * Check if wallet is locked
   */
  async isLocked(): Promise<boolean> {
    const response = await this.api.get('/api/v1/wallet/locked');
    return response.data.locked;
  }

  /**
   * Change wallet passphrase
   */
  async changePassphrase(oldPassphrase: string, newPassphrase: string): Promise<void> {
    await this.api.post('/api/v1/wallet/change-passphrase', {
      oldPassphrase,
      newPassphrase
    });
  }

  /**
   * Backup wallet
   */
  async backupWallet(passphrase?: string): Promise<string> {
    const response = await this.api.post('/api/v1/wallet/backup', {
      passphrase
    });

    return response.data.backup;
  }

  /**
   * Restore wallet from backup
   */
  async restoreWallet(backup: string, passphrase?: string): Promise<Account> {
    const response = await this.api.post('/api/v1/wallet/restore', {
      backup,
      passphrase
    });

    return response.data;
  }
}
