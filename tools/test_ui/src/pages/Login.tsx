import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { request, ApiError } from '../api/client'

export default function Login() {
  const { login, token, user } = useAuth()
  const [email, setEmail] = useState('admin@example.com')
  const [password, setPassword] = useState('password')
  const [username, setUsername] = useState('')
  const [fullName, setFullName] = useState('')
  const [error, setError] = useState('')
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const navigate = useNavigate()

  const handleLogin = async () => {
    setError('')
    try { await login(email, password); navigate('/') }
    catch (e) { setError(e instanceof ApiError ? e.message : String(e)) }
  }

  const handleRegister = async () => {
    setError('')
    try {
      const data = await request<any>('POST', '/auth/register', { email, password, username, fullName })
      if (data.accessToken) {
        localStorage.setItem('innogen_token', data.accessToken)
        window.location.href = '/'
      }
    } catch (e) { setError(e instanceof ApiError ? e.message : String(e)) }
  }

  if (token) return <div className="card"><h3>Logged In</h3><pre>{JSON.stringify(user, null, 2)}</pre></div>

  return (
    <div className="card" style={{ maxWidth: 400 }}>
      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
        <button onClick={() => setMode('login')} style={{ background: mode === 'login' ? '#4a6cf7' : '#ccc' }}>Login</button>
        <button onClick={() => setMode('register')} style={{ background: mode === 'register' ? '#4a6cf7' : '#ccc' }}>Register</button>
      </div>

      <h3>{mode === 'login' ? 'Login' : 'Register'}</h3>

      {mode === 'register' && (
        <>
          <div className="form-group"><label>Username</label><input value={username} onChange={e => setUsername(e.target.value)} /></div>
          <div className="form-group"><label>Full Name</label><input value={fullName} onChange={e => setFullName(e.target.value)} /></div>
        </>
      )}

      <div className="form-group"><label>Email</label><input value={email} onChange={e => setEmail(e.target.value)} /></div>
      <div className="form-group"><label>Password</label><input type="password" value={password} onChange={e => setPassword(e.target.value)} /></div>

      {error && <p className="error">{error}</p>}

      {mode === 'login' ? (
        <button onClick={handleLogin}>Login</button>
      ) : (
        <button onClick={handleRegister}>Register</button>
      )}
    </div>
  )
}
