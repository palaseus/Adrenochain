import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { WalletProvider } from './hooks/useWallet'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Send from './pages/Send'
import Receive from './pages/Receive'
import DeFi from './pages/DeFi'
import Settings from './pages/Settings'
import Login from './pages/Login'
import CreateWallet from './pages/CreateWallet'
import ImportWallet from './pages/ImportWallet'

function App() {
  return (
    <WalletProvider>
      <div className="min-h-screen bg-gray-50">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/create" element={<CreateWallet />} />
          <Route path="/import" element={<ImportWallet />} />
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="send" element={<Send />} />
            <Route path="receive" element={<Receive />} />
            <Route path="defi" element={<DeFi />} />
            <Route path="settings" element={<Settings />} />
          </Route>
        </Routes>
      </div>
    </WalletProvider>
  )
}

export default App
