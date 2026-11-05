import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminTags() {
  const [name, setName] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/tags', { name })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/tags'); setList(d.tags || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Tags</h3>
      <div className="form-group"><label>Name</label><input value={name} onChange={e => setName(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
