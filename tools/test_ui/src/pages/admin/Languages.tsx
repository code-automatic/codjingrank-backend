import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminLanguages() {
  const [name, setName] = useState('')
  const [pistonAlias, setPistonAlias] = useState('')
  const [pistonVersion, setPistonVersion] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/languages', { name, pistonAlias, pistonVersion })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/languages'); setList(d.languages || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Languages</h3>
      <div className="form-group"><label>Name</label><input value={name} onChange={e => setName(e.target.value)} /></div>
      <div className="form-group"><label>Piston Alias</label><input value={pistonAlias} onChange={e => setPistonAlias(e.target.value)} /></div>
      <div className="form-group"><label>Piston Version</label><input value={pistonVersion} onChange={e => setPistonVersion(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
