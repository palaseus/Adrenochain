import React, { useEffect } from 'react'
import { useWallet } from '../hooks/useWallet'
import { Wallet, Send, Download, TrendingUp, Activity, Shield } from 'lucide-react'

const Dashboard: React.FC = () => {
  const { address, balance, isLoading, getBalance } = useWallet()

  useEffect(() => {
    if (address) {
      getBalance()
    }
  }, [address, getBalance])

  const formatAddress = (addr: string) => {
    return `${addr.slice(0, 6)}...${addr.slice(-4)}`
  }

  const formatBalance = (bal: string) => {
    return parseFloat(bal).toFixed(6)
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-2 text-gray-600">Welcome to your Adrenochain wallet</p>
      </div>

      {/* Wallet Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* Balance Card */}
        <div className="lg:col-span-2">
          <div className="card">
            <div className="card-body">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-medium text-gray-900">Wallet Balance</h2>
                <Wallet className="h-6 w-6 text-blue-600" />
              </div>
              
              <div className="text-center py-8">
                <div className="text-4xl font-bold text-gray-900 mb-2">
                  {isLoading ? (
                    <div className="animate-pulse bg-gray-200 h-12 w-32 mx-auto rounded"></div>
                  ) : (
                    `${formatBalance(balance)} ADR`
                  )}
                </div>
                <p className="text-gray-600">
                  {address ? formatAddress(address) : 'No wallet connected'}
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4 mt-6">
                <button className="btn-primary flex items-center justify-center">
                  <Send className="h-4 w-4 mr-2" />
                  Send
                </button>
                <button className="btn-secondary flex items-center justify-center">
                  <Download className="h-4 w-4 mr-2" />
                  Receive
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="space-y-6">
          <div className="card">
            <div className="card-body">
              <div className="flex items-center">
                <TrendingUp className="h-8 w-8 text-green-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Portfolio Value</p>
                  <p className="text-2xl font-bold text-gray-900">$1,234.56</p>
                </div>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="card-body">
              <div className="flex items-center">
                <Activity className="h-8 w-8 text-blue-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">24h Change</p>
                  <p className="text-2xl font-bold text-green-600">+5.2%</p>
                </div>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="card-body">
              <div className="flex items-center">
                <Shield className="h-8 w-8 text-purple-600" />
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600">Security Score</p>
                  <p className="text-2xl font-bold text-gray-900">98%</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Transactions */}
        <div className="card">
          <div className="card-header">
            <h3 className="text-lg font-medium text-gray-900">Recent Transactions</h3>
          </div>
          <div className="card-body">
            <div className="space-y-4">
              {[
                { type: 'send', to: '0x1234...5678', amount: '10.5 ADR', time: '2 hours ago', status: 'confirmed' },
                { type: 'receive', from: '0x8765...4321', amount: '25.0 ADR', time: '5 hours ago', status: 'confirmed' },
                { type: 'send', to: '0xabcd...efgh', amount: '5.2 ADR', time: '1 day ago', status: 'pending' },
              ].map((tx, index) => (
                <div key={index} className="flex items-center justify-between py-3 border-b border-gray-200 last:border-b-0">
                  <div className="flex items-center">
                    <div className={`w-2 h-2 rounded-full mr-3 ${
                      tx.type === 'send' ? 'bg-red-500' : 'bg-green-500'
                    }`} />
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {tx.type === 'send' ? 'Sent to' : 'Received from'} {tx.type === 'send' ? tx.to : tx.from}
                      </p>
                      <p className="text-xs text-gray-500">{tx.time}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={`text-sm font-medium ${
                      tx.type === 'send' ? 'text-red-600' : 'text-green-600'
                    }`}>
                      {tx.type === 'send' ? '-' : '+'}{tx.amount}
                    </p>
                    <p className={`text-xs ${
                      tx.status === 'confirmed' ? 'text-green-600' : 'text-yellow-600'
                    }`}>
                      {tx.status}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* DeFi Opportunities */}
        <div className="card">
          <div className="card-header">
            <h3 className="text-lg font-medium text-gray-900">DeFi Opportunities</h3>
          </div>
          <div className="card-body">
            <div className="space-y-4">
              {[
                { name: 'ADR/USDC Pool', apy: '12.5%', tvl: '$2.1M', risk: 'Low' },
                { name: 'ADR/ETH Pool', apy: '18.3%', tvl: '$1.8M', risk: 'Medium' },
                { name: 'Staking Pool', apy: '8.7%', tvl: '$5.2M', risk: 'Low' },
              ].map((pool, index) => (
                <div key={index} className="flex items-center justify-between py-3 border-b border-gray-200 last:border-b-0">
                  <div>
                    <p className="text-sm font-medium text-gray-900">{pool.name}</p>
                    <p className="text-xs text-gray-500">TVL: {pool.tvl}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium text-green-600">{pool.apy} APY</p>
                    <p className={`text-xs ${
                      pool.risk === 'Low' ? 'text-green-600' : 
                      pool.risk === 'Medium' ? 'text-yellow-600' : 'text-red-600'
                    }`}>
                      {pool.risk} Risk
                    </p>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4">
              <button className="btn-primary w-full">View All Pools</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
