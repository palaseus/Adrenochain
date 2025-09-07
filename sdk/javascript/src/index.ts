/**
 * Adrenochain JavaScript/TypeScript SDK
 * 
 * This SDK provides a comprehensive interface for interacting with the Adrenochain blockchain,
 * including wallet management, transaction creation, DeFi operations, and more.
 */

import { AdrenochainClient } from './client';
import { Wallet } from './wallet';
import { Transaction } from './transaction';
import { DeFi } from './defi';
import { Explorer } from './explorer';
import { types } from './types';

// Export main classes
export {
  AdrenochainClient,
  Wallet,
  Transaction,
  DeFi,
  Explorer,
  types
};

// Export types
export * from './types';

// Default export
export default AdrenochainClient;
