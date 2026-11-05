import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminSessions() {
  const [subjectId, setSubjectId] = useState('')
  const [title, setTitle] = useState('')
  const [orderIndex, setOrderIndex] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/sessions', { subjectId: +subjectId, title, orderIndex: +orderIndex })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/sessions'); setList(d.sessions || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Sessions</h3>
      <div className="form-group"><label>Subject ID</label><input value={subjectId} onChange={e => setSubjectId(e.target.value)} /></div>
      <div className="form-group"><label>Title</label><input value={title} onChange={e => setTitle(e.target.value)} /></div>
      <div className="form-group"><label>Order Index</label><input value={orderIndex} onChange={e => setOrderIndex(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
