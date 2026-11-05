import { Link, Outlet } from 'react-router-dom'
export default function AdminLayout() {
  const tabs = ['languages', 'subjects', 'sessions', 'lessons', 'problems', 'test-cases', 'tags']
  return (
    <div>
      <h2>Admin</h2>
      <div className="sidebar">
        {tabs.map(t => <Link key={t} to={`/admin/${t}`} style={{ padding: '4px 12px', background: '#eee', borderRadius: 4, textDecoration: 'none', color: '#333' }}>{t}</Link>)}
      </div>
      <Outlet />
    </div>
  )
}
