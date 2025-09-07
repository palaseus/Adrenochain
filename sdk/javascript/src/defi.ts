import { AxiosInstance } from 'axios';
import { DeFiConfig, DeFiPool, DeFiPosition, GovernanceProposal } from './types';

/**
 * DeFi class for interacting with decentralized finance protocols
 */
export class DeFi {
  constructor(private api: AxiosInstance) {}

  // ===== LIQUIDITY POOLS =====

  /**
   * Get all available DeFi pools
   */
  async getPools(): Promise<DeFiPool[]> {
    const response = await this.api.get('/api/v1/defi/pools');
    return response.data.pools;
  }

  /**
   * Get pool by ID
   */
  async getPool(poolId: string): Promise<DeFiPool> {
    const response = await this.api.get(`/api/v1/defi/pools/${poolId}`);
    return response.data;
  }

  /**
   * Create a new liquidity pool
   */
  async createPool(tokenA: string, tokenB: string, fee: number): Promise<DeFiPool> {
    const response = await this.api.post('/api/v1/defi/pools', {
      tokenA,
      tokenB,
      fee
    });
    return response.data;
  }

  /**
   * Add liquidity to a pool
   */
  async addLiquidity(config: DeFiConfig): Promise<DeFiPosition> {
    const response = await this.api.post('/api/v1/defi/liquidity/add', config);
    return response.data;
  }

  /**
   * Remove liquidity from a pool
   */
  async removeLiquidity(positionId: string, amount: string): Promise<{
    tokenA: string;
    tokenB: string;
    txHash: string;
  }> {
    const response = await this.api.post('/api/v1/defi/liquidity/remove', {
      positionId,
      amount
    });
    return response.data;
  }

  /**
   * Get liquidity positions for an address
   */
  async getPositions(address: string): Promise<DeFiPosition[]> {
    const response = await this.api.get(`/api/v1/defi/positions/${address}`);
    return response.data.positions;
  }

  /**
   * Get position by ID
   */
  async getPosition(positionId: string): Promise<DeFiPosition> {
    const response = await this.api.get(`/api/v1/defi/positions/${positionId}`);
    return response.data;
  }

  // ===== SWAPPING =====

  /**
   * Get swap quote
   */
  async getSwapQuote(tokenIn: string, tokenOut: string, amountIn: string): Promise<{
    amountOut: string;
    priceImpact: number;
    route: string[];
    fee: string;
  }> {
    const response = await this.api.get('/api/v1/defi/swap/quote', {
      params: { tokenIn, tokenOut, amountIn }
    });
    return response.data;
  }

  /**
   * Execute a swap
   */
  async swap(tokenIn: string, tokenOut: string, amountIn: string, minAmountOut: string, recipient?: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/swap', {
      tokenIn,
      tokenOut,
      amountIn,
      minAmountOut,
      recipient
    });
    return response.data.txHash;
  }

  // ===== LENDING =====

  /**
   * Get lending markets
   */
  async getLendingMarkets(): Promise<{
    markets: Array<{
      id: string;
      token: string;
      totalSupply: string;
      totalBorrow: string;
      supplyRate: number;
      borrowRate: number;
      collateralFactor: number;
    }>;
  }> {
    const response = await this.api.get('/api/v1/defi/lending/markets');
    return response.data;
  }

  /**
   * Supply assets to lending protocol
   */
  async supply(token: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/lending/supply', {
      token,
      amount
    });
    return response.data.txHash;
  }

  /**
   * Borrow assets from lending protocol
   */
  async borrow(token: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/lending/borrow', {
      token,
      amount
    });
    return response.data.txHash;
  }

  /**
   * Repay borrowed assets
   */
  async repay(token: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/lending/repay', {
      token,
      amount
    });
    return response.data.txHash;
  }

  /**
   * Withdraw supplied assets
   */
  async withdraw(token: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/lending/withdraw', {
      token,
      amount
    });
    return response.data.txHash;
  }

  // ===== STAKING =====

  /**
   * Get staking pools
   */
  async getStakingPools(): Promise<{
    pools: Array<{
      id: string;
      token: string;
      totalStaked: string;
      apy: number;
      lockPeriod: number;
      minStake: string;
    }>;
  }> {
    const response = await this.api.get('/api/v1/defi/staking/pools');
    return response.data;
  }

  /**
   * Stake tokens
   */
  async stake(poolId: string, amount: string, lockPeriod?: number): Promise<string> {
    const response = await this.api.post('/api/v1/defi/staking/stake', {
      poolId,
      amount,
      lockPeriod
    });
    return response.data.txHash;
  }

  /**
   * Unstake tokens
   */
  async unstake(poolId: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/staking/unstake', {
      poolId,
      amount
    });
    return response.data.txHash;
  }

  /**
   * Claim staking rewards
   */
  async claimRewards(poolId: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/staking/claim', {
      poolId
    });
    return response.data.txHash;
  }

  // ===== YIELD FARMING =====

  /**
   * Get yield farming pools
   */
  async getYieldPools(): Promise<{
    pools: Array<{
      id: string;
      name: string;
      token: string;
      rewardToken: string;
      totalStaked: string;
      apy: number;
      tvl: string;
    }>;
  }> {
    const response = await this.api.get('/api/v1/defi/yield/pools');
    return response.data;
  }

  /**
   * Deposit into yield farm
   */
  async depositYield(poolId: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/yield/deposit', {
      poolId,
      amount
    });
    return response.data.txHash;
  }

  /**
   * Withdraw from yield farm
   */
  async withdrawYield(poolId: string, amount: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/yield/withdraw', {
      poolId,
      amount
    });
    return response.data.txHash;
  }

  /**
   * Harvest yield rewards
   */
  async harvestYield(poolId: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/yield/harvest', {
      poolId
    });
    return response.data.txHash;
  }

  // ===== GOVERNANCE =====

  /**
   * Get governance proposals
   */
  async getProposals(): Promise<GovernanceProposal[]> {
    const response = await this.api.get('/api/v1/defi/governance/proposals');
    return response.data.proposals;
  }

  /**
   * Get proposal by ID
   */
  async getProposal(proposalId: number): Promise<GovernanceProposal> {
    const response = await this.api.get(`/api/v1/defi/governance/proposals/${proposalId}`);
    return response.data;
  }

  /**
   * Create a governance proposal
   */
  async createProposal(title: string, description: string, actions: any[]): Promise<number> {
    const response = await this.api.post('/api/v1/defi/governance/proposals', {
      title,
      description,
      actions
    });
    return response.data.proposalId;
  }

  /**
   * Vote on a proposal
   */
  async vote(proposalId: number, support: 'for' | 'against' | 'abstain', reason?: string): Promise<string> {
    const response = await this.api.post('/api/v1/defi/governance/vote', {
      proposalId,
      support,
      reason
    });
    return response.data.txHash;
  }

  /**
   * Execute a proposal
   */
  async executeProposal(proposalId: number): Promise<string> {
    const response = await this.api.post('/api/v1/defi/governance/execute', {
      proposalId
    });
    return response.data.txHash;
  }

  // ===== ANALYTICS =====

  /**
   * Get DeFi analytics
   */
  async getAnalytics(): Promise<{
    totalValueLocked: string;
    totalVolume24h: string;
    totalFees24h: string;
    activeUsers: number;
    topPools: DeFiPool[];
  }> {
    const response = await this.api.get('/api/v1/defi/analytics');
    return response.data;
  }

  /**
   * Get portfolio analytics for an address
   */
  async getPortfolioAnalytics(address: string): Promise<{
    totalValue: string;
    positions: DeFiPosition[];
    pnl24h: string;
    apy: number;
  }> {
    const response = await this.api.get(`/api/v1/defi/portfolio/${address}`);
    return response.data;
  }
}
