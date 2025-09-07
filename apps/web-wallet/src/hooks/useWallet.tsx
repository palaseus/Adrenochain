import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import toast from 'react-hot-toast'

interface WalletContextType {
  isConnected: boolean
  address: string | null
  balance: string
  connect: (privateKey: string, passphrase?: string) => Promise<void>
  disconnect: () => void
  sendTransaction: (to: string, amount: string, fee?: string) => Promise<string>
  getBalance: () => Promise<void>
  isLoading: boolean
}

const WalletContext = createContext<WalletContextType | undefined>(undefined)

export const useWallet = () => {
  const context = useContext(WalletContext)
  if (context === undefined) {
    throw new Error('useWallet must be used within a WalletProvider')
  }
  return context
}

interface WalletProviderProps {
  children: ReactNode
}

export const WalletProvider: React.FC<WalletProviderProps> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false)
  const [address, setAddress] = useState<string | null>(null)
  const [balance, setBalance] = useState('0')
  const [isLoading, setIsLoading] = useState(false)

  // Check for existing wallet on mount
  useEffect(() => {
    const savedWallet = localStorage.getItem('adrenochain_wallet')
    if (savedWallet) {
      try {
        const walletData = JSON.parse(savedWallet)
        setAddress(walletData.address)
        setIsConnected(true)
        getBalance()
      } catch (error) {
        console.error('Failed to load saved wallet:', error)
        localStorage.removeItem('adrenochain_wallet')
      }
    }
  }, [])

  const connect = async (privateKey: string, passphrase?: string) => {
    setIsLoading(true)
    try {
      // In a real implementation, this would use the Adrenochain SDK
      // For now, we'll simulate wallet connection
      const mockAddress = '0x' + privateKey.slice(0, 40)
      
      setAddress(mockAddress)
      setIsConnected(true)
      
      // Save wallet to localStorage (in production, use secure storage)
      localStorage.setItem('adrenochain_wallet', JSON.stringify({
        address: mockAddress,
        privateKey: privateKey, // In production, encrypt this
        timestamp: Date.now()
      }))
      
      await getBalance()
      toast.success('Wallet connected successfully!')
    } catch (error) {
      console.error('Failed to connect wallet:', error)
      toast.error('Failed to connect wallet')
    } finally {
      setIsLoading(false)
    }
  }

  const disconnect = () => {
    setIsConnected(false)
    setAddress(null)
    setBalance('0')
    localStorage.removeItem('adrenochain_wallet')
    toast.success('Wallet disconnected')
  }

  const getBalance = async () => {
    if (!address) return
    
    setIsLoading(true)
    try {
      // In a real implementation, this would call the Adrenochain API
      // For now, we'll simulate a balance
      const mockBalance = (Math.random() * 1000).toFixed(6)
      setBalance(mockBalance)
    } catch (error) {
      console.error('Failed to get balance:', error)
      toast.error('Failed to fetch balance')
    } finally {
      setIsLoading(false)
    }
  }

  const sendTransaction = async (to: string, amount: string, fee?: string): Promise<string> => {
    if (!address) {
      throw new Error('Wallet not connected')
    }

    setIsLoading(true)
    try {
      // In a real implementation, this would use the Adrenochain SDK
      // For now, we'll simulate a transaction
      const mockTxHash = '0x' + Math.random().toString(16).substr(2, 64)
      
      // Simulate transaction processing
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      // Update balance
      const newBalance = (parseFloat(balance) - parseFloat(amount)).toFixed(6)
      setBalance(newBalance)
      
      toast.success('Transaction sent successfully!')
      return mockTxHash
    } catch (error) {
      console.error('Failed to send transaction:', error)
      toast.error('Failed to send transaction')
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const value: WalletContextType = {
    isConnected,
    address,
    balance,
    connect,
    disconnect,
    sendTransaction,
    getBalance,
    isLoading
  }

  return (
    <WalletContext.Provider value={value}>
      {children}
    </WalletContext.Provider>
  )
}
