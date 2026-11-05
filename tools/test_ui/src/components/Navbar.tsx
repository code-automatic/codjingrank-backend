import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

export default function Navbar() {
  const { token, user, logout } = useAuth()
  const navigate = useNavigate()
  return (
    <nav>
      <Link to="/" className="brand">InnoGen Test UI</Link>
      <Link to="/curriculum">Curriculum</Link>
      <Link to="/runner">Runner</Link>
      <Link to="/submissions">Submissions</Link>
      {user?.role === 'admin' && <Link to="/admin">Admin</Link>}
      <Link to="/repo-notes">GitHub/Webhook</Link>
      <Link to="/github/connect">GitHub</Link>
      {token ? <button onClick={() => { logout(); navigate('/') }}>Logout</button> : <Link to="/login">Login</Link>}
    </nav>
  )
}
