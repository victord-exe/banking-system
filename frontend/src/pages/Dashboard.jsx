import { useAuth } from '../context/AuthContext'
import { HiCreditCard } from 'react-icons/hi'
import { RiShieldStarFill } from 'react-icons/ri'

const Dashboard = () => {
  const { user, balance, balanceError } = useAuth()

  return (
    <div className="page-container">
      <div className="balance">
        <h3><HiCreditCard className="inline-icon" /> Account Balance</h3>
        {balanceError ? (
          <div className="error" style={{ marginTop: '1rem' }}>{balanceError}</div>
        ) : (
          <>
            <div className="amount">${balance !== null ? balance.toFixed(2) : '0.00'}</div>
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
