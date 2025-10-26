import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { HiSparkles } from 'react-icons/hi'

const Login = () => {
  const [isLogin, setIsLogin] = useState(true)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [fullName, setFullName] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)

  const { login, register } = useAuth()
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    setLoading(true)

    try {
      if (isLogin) {
        await login(email, password)
        setSuccess('Login successful!')
      } else {
        await register(email, password, fullName)
        setSuccess('Registration successful!')
      }

      // Redirect to dashboard after successful login/register
      setTimeout(() => {
        navigate('/dashboard')
      }, 500)
    } catch (err) {
      setError(err.response?.data?.message || err.response?.data?.error || 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container">
      <div className="card">
        <div className="logo-area">
          <div className="tech-icon">
            <HiSparkles size={48} className="sparkle-icon" />
          </div>
        </div>
        <h1>HLABS Banking</h1>
        <h2>{isLogin ? 'Welcome Back' : 'Create Account'}</h2>

        {error && <div className="error">{error}</div>}
        {success && <div className="success">{success}</div>}

        <form onSubmit={handleSubmit}>
          {!isLogin && (
            <div className="form-group">
              <label htmlFor="fullName">Full Name</label>
              <input
                type="text"
                id="fullName"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required={!isLogin}
                placeholder="John Doe"
              />
            </div>
          )}

          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="you@example.com"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              placeholder="At least 6 characters"
            />
          </div>

          <button type="submit" disabled={loading}>
            {loading ? 'Please wait...' : isLogin ? 'Sign In' : 'Sign Up'}
          </button>
        </form>

        <div className="toggle-link">
          {isLogin ? (
            <>
              Don't have an account?{' '}
              <a onClick={() => setIsLogin(false)}>Sign up</a>
            </>
          ) : (
            <>
              Already have an account?{' '}
              <a onClick={() => setIsLogin(true)}>Sign in</a>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export default Login
