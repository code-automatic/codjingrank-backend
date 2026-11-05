import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminLessons() {
  const [sessionId, setSessionId] = useState('')
  const [title, setTitle] = useState('')
  const [orderIndex, setOrderIndex] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/lessons', { sessionId: +sessionId, title, orderIndex: +orderIndex })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/lessons'); setList(d.lessons || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Lessons</h3>
      <div className="form-group"><label>Session ID</label><input value={sessionId} onChange={e => setSessionId(e.target.value)} /></div>
      <div className="form-group"><label>Title</label><input value={title} onChange={e => setTitle(e.target.value)} /></div>
      <div className="form-group"><label>Order Index</label><input value={orderIndex} onChange={e => setOrderIndex(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
