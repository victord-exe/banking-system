import { useEffect } from 'react'
import { useAuth } from '../context/AuthContext'
import { HiCreditCard } from 'react-icons/hi'
import { RiShieldStarFill } from 'react-icons/ri'

const Dashboard = () => {
  const { user, balance, balanceError, fetchBalance, loading } = useAuth()

  console.log('🟢 [Dashboard] Rendered with:', { user: user?.email, balance, balanceError, loading })

  // Auto-load balance on mount + polling every 30 seconds
  useEffect(() => {
    console.log('🟢 [Dashboard] useEffect - Component mounted, loading initial balance...')
    fetchBalance() // Initial load

    // Set up polling interval
    console.log('🟢 [Dashboard] Setting up polling interval (30s)...')
    const interval = setInterval(() => {
      console.log('🟢 [Dashboard] Polling interval triggered - fetching balance...')
      fetchBalance()
    }, 30000) // 30 seconds

    // Cleanup interval on unmount
    return () => {
      console.log('🟢 [Dashboard] Component unmounting - cleaning up polling interval')
      clearInterval(interval)
    }
  }, [fetchBalance])

  const formatBalance = (value) => {
    if (value === null || value === undefined) return '0.00'
    return value.toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  }

  return (
    <div className="page-container">
      <div className="balance">
        <h3><HiCreditCard className="inline-icon" /> Account Balance</h3>
        {balanceError ? (
          <div className="error" style={{ marginTop: '1rem' }}>{balanceError}</div>
        ) : (
          <>
            <div className="amount" style={{ fontSize: '3rem', fontWeight: 'bold', marginTop: '1rem' }}>
              ${formatBalance(balance)}
            </div>
            <div className="balance-currency" style={{ fontSize: '1rem', color: 'rgba(255, 255, 255, 0.6)', marginTop: '0.5rem' }}>
              United States Dollar (USD)
            </div>
            <div className="balance-indicator">
              <span className="status-dot"></span>
              <span className="status-text">Active • Secure</span>
            </div>
          </>
        )}
      </div>

      <div className="card">
        <div className="card-header">
          <h2>Account Information</h2>
          <div className="tech-badge">
            <RiShieldStarFill />
            <span>Verified</span>
          </div>
        </div>
        <div className="user-info">
          <p><strong>Name:</strong> <span>{user?.full_name}</span></p>
          <p><strong>Email:</strong> <span>{user?.email}</span></p>
          <p><strong>Account ID:</strong> <span className="mono">{user?.tigerbeetle_account_id}</span></p>
          <p><strong>User ID:</strong> <span className="mono">{user?.id}</span></p>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
