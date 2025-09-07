import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useWallet } from '../hooks/useWallet'
import { Wallet, Lock, Eye, EyeOff } from 'lucide-react'

const Login: React.FC = () => {
  const [privateKey, setPrivateKey] = useState('')
  const [passphrase, setPassphrase] = useState('')
  const [showPassphrase, setShowPassphrase] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const { connect } = useWallet()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!privateKey.trim()) {
      return
    }

    setIsLoading(true)
    try {
      await connect(privateKey, passphrase || undefined)
      navigate('/')
    } catch (error) {
      console.error('Login failed:', error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100">
            <Wallet className="h-6 w-6 text-blue-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your wallet
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{' '}
            <Link
              to="/create"
              className="font-medium text-blue-600 hover:text-blue-500"
            >
              create a new wallet
            </Link>
          </p>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label htmlFor="privateKey" className="block text-sm font-medium text-gray-700">
                Private Key
              </label>
              <div className="mt-1">
                <textarea
                  id="privateKey"
                  name="privateKey"
                  rows={3}
                  required
                  className="input"
                  placeholder="Enter your private key (0x...)"
                  value={privateKey}
                  onChange={(e) => setPrivateKey(e.target.value)}
                />
              </div>
            </div>

            <div>
              <label htmlFor="passphrase" className="block text-sm font-medium text-gray-700">
                Passphrase (Optional)
              </label>
              <div className="mt-1 relative">
                <input
                  id="passphrase"
                  name="passphrase"
                  type={showPassphrase ? 'text' : 'password'}
                  className="input pr-10"
                  placeholder="Enter passphrase if wallet is encrypted"
                  value={passphrase}
                  onChange={(e) => setPassphrase(e.target.value)}
                />
                <button
                  type="button"
                  className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  onClick={() => setShowPassphrase(!showPassphrase)}
                >
                  {showPassphrase ? (
                    <EyeOff className="h-5 w-5 text-gray-400" />
                  ) : (
                    <Eye className="h-5 w-5 text-gray-400" />
                  )}
                </button>
              </div>
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={isLoading || !privateKey.trim()}
              className="btn-primary w-full flex justify-center items-center"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Connecting...
                </>
              ) : (
                <>
                  <Lock className="h-4 w-4 mr-2" />
                  Connect Wallet
                </>
              )}
            </button>
          </div>

          <div className="text-center">
            <Link
              to="/import"
              className="text-sm text-blue-600 hover:text-blue-500"
            >
              Import wallet from mnemonic
            </Link>
          </div>
        </form>

        <div className="mt-8">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-gray-50 text-gray-500">Security Notice</span>
            </div>
          </div>
          <div className="mt-4 text-xs text-gray-500 text-center">
            <p>
              Your private key is stored locally and never transmitted to our servers.
              Always ensure you're using a secure connection.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Login
