import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { HiSparkles, HiHome, HiCreditCard, HiClock, HiChat, HiLogout } from 'react-icons/hi'

const Navbar = () => {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <div className="navbar-brand">
          <HiSparkles className="navbar-logo-icon" />
          <span className="navbar-title">HLABS Banking</span>
        </div>

        <div className="navbar-links">
          <NavLink to="/dashboard" className="nav-link">
            <HiHome className="nav-icon" />
            <span>Dashboard</span>
          </NavLink>

          <NavLink to="/transactions" className="nav-link">
            <HiCreditCard className="nav-icon" />
            <span>Transactions</span>
          </NavLink>

          <NavLink to="/history" className="nav-link">
            <HiClock className="nav-icon" />
            <span>History</span>
          </NavLink>

          <NavLink to="/chat" className="nav-link">
            <HiChat className="nav-icon" />
            <span>AI Chat</span>
          </NavLink>
        </div>

        <div className="navbar-user">
          <div className="user-info-nav">
            <span className="user-name">{user?.full_name}</span>
            <span className="user-email">{user?.email}</span>
          </div>
          <button onClick={handleLogout} className="logout-btn-nav">
            <HiLogout />
          </button>
        </div>
      </div>
    </nav>
  )
}

export default Navbar
