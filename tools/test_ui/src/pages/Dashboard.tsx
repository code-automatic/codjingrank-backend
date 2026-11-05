import { useState } from 'react'
import { request } from '../api/client'
import { useConfig } from '../hooks/useConfig'
import { useAuth } from '../hooks/useAuth'

export default function Dashboard() {
  const { baseURL, setURL } = useConfig()
  const { token, user } = useAuth()
  const [status, setStatus] = useState<Record<string, string>>({})
  const [directStatus, setDirectStatus] = useState<Record<string, string>>({})

  const checkGateway = async () => {
    try {
      const d = await request<any>('GET', '/health')
      setStatus(s => ({ ...s, gateway: `OK (${d.service || 'api-gateway'})` }))
    } catch { setStatus(s => ({ ...s, gateway: 'FAIL' })) }
  }

  const checkDirect = async (name: string, path: string) => {
    try {
      const d = await request<any>('GET', path)
      setDirectStatus(s => ({ ...s, [name]: `OK (${d.service || name})` }))
    } catch { setDirectStatus(s => ({ ...s, [name]: 'FAIL' })) }
  }

  return (
    <div>
      <h2>Dashboard</h2>
      <div className="card">
        <h3>API Base URL</h3>
        <div className="form-group">
          <input value={baseURL} onChange={e => setURL(e.target.value)} />
        </div>
        <button onClick={checkGateway}>Check Gateway Health</button>
        {status.gateway && <p className={status.gateway.startsWith('OK') ? 'success' : 'error'}>{status.gateway}</p>}
      </div>

      <div className="card">
        <h3>Direct Service Health</h3>
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          {[
            ['auth', '/health/auth'],
            ['run', '/health/run'],
            ['submission', '/health/submission'],
            ['repo', '/health/repo'],
          ].map(([name, path]) => (
            <button key={name} onClick={() => checkDirect(name, path)}>{name}</button>
          ))}
        </div>
        {Object.entries(directStatus).map(([k, v]) => (
          <p key={k} className={v.startsWith('OK') ? 'success' : 'error'}>{k}: {v}</p>
        ))}
      </div>

      {user && (
        <div className="card">
          <h3>Logged In</h3>
          <pre>{JSON.stringify(user, null, 2)}</pre>
        </div>
      )}

      {user && (
        <div className="card">
          <h3>GitHub App</h3>
          <button onClick={async () => {
            try {
              const d = await request<any>('GET', '/auth/github/connect')
              if (d.installUrl) window.location.href = d.installUrl
            } catch(e: any) { alert('Failed: ' + e.message) }
          }}>Connect GitHub App</button>
          <p style={{fontSize:'0.85rem', marginTop:'0.5rem', color:'#666'}}>
            Installs the RinnoGen GitHub App to your personal account or organization.
          </p>
        </div>
      )}
    </div>
  )
}
