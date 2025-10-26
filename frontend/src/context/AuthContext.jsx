import { createContext, useContext, useState, useEffect } from 'react'
import { authAPI, accountAPI } from '../services/api'

const AuthContext = createContext(null)

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [balance, setBalance] = useState(null)
  const [loading, setLoading] = useState(true)
  const [balanceError, setBalanceError] = useState('')

  useEffect(() => {
    // Check if user is already logged in
    const token = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')

    if (token && savedUser) {
      setUser(JSON.parse(savedUser))
      fetchBalance()
    }
    setLoading(false)
  }, [])

  const fetchBalance = async () => {
    try {
      setBalanceError('')
      const response = await accountAPI.getBalance()
      setBalance(response.data.balance || 0)
    } catch (err) {
      console.error('Failed to fetch balance:', err)
      setBalanceError('Unable to fetch balance')
      setBalance(0)
    }
  }

  const login = async (email, password) => {
    const response = await authAPI.login({ email, password })
    const { token, user: userData } = response.data.data

    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(userData))

    setUser(userData)
    await fetchBalance()

    return response
  }

  const register = async (email, password, fullName) => {
    const response = await authAPI.register({
      email,
      password,
      full_name: fullName
    })
    const { token, user: userData } = response.data.data

    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(userData))

    setUser(userData)
    await fetchBalance()

    return response
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setUser(null)
    setBalance(null)
    setBalanceError('')
  }

  const value = {
    user,
    balance,
    balanceError,
    loading,
    login,
    register,
    logout,
    fetchBalance,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
