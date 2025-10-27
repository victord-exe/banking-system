import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { authAPI, accountAPI } from '../services/api'

const AuthContext = createContext(null)

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [balance, setBalance] = useState(null)
  const [loading, setLoading] = useState(true)
  const [balanceError, setBalanceError] = useState('')

  const fetchBalance = useCallback(async () => {
    console.log('🔵 [AuthContext] fetchBalance called')
    try {
      setBalanceError('')
      console.log('🔵 [AuthContext] Calling accountAPI.getBalance()...')
      const response = await accountAPI.getBalance()
      console.log('🔵 [AuthContext] API response received:', response)
      console.log('🔵 [AuthContext] response.data:', response.data)

      // Backend returns balance in cents, convert to dollars
      // Backend structure: { data: { balance: X, currency: "USD" }, message: "..." }
      const balanceInCents = response.data.data.balance || 0
      const balanceInDollars = balanceInCents / 100
      console.log(`🔵 [AuthContext] Balance in cents: ${balanceInCents}, Balance in dollars: ${balanceInDollars}`)

      setBalance(balanceInDollars)
      console.log('✅ [AuthContext] Balance updated successfully:', balanceInDollars)
    } catch (err) {
      console.error('❌ [AuthContext] Failed to fetch balance:', err)
      console.error('❌ [AuthContext] Error details:', {
        message: err.message,
        response: err.response,
        status: err.response?.status,
        data: err.response?.data
      })
      setBalanceError('Unable to fetch balance')
      setBalance(0)
    }
  }, [])

  useEffect(() => {
    console.log('🟡 [AuthContext] useEffect - Checking for existing session...')
    // Check if user is already logged in
    const token = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    console.log('🟡 [AuthContext] Token exists:', !!token)
    console.log('🟡 [AuthContext] SavedUser exists:', !!savedUser)

    if (token && savedUser) {
      const parsedUser = JSON.parse(savedUser)
      console.log('🟡 [AuthContext] User restored from localStorage:', parsedUser)
      setUser(parsedUser)
      console.log('🟡 [AuthContext] Fetching initial balance...')
      fetchBalance()
    } else {
      console.log('🟡 [AuthContext] No existing session found')
    }
    setLoading(false)
  }, [fetchBalance])

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
