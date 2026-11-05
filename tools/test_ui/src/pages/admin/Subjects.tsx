import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminSubjects() {
  const [title, setTitle] = useState('')
  const [slug, setSlug] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/subjects', { title, slug })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/subjects'); setList(d.subjects || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Subjects</h3>
      <div className="form-group"><label>Title</label><input value={title} onChange={e => setTitle(e.target.value)} /></div>
      <div className="form-group"><label>Slug</label><input value={slug} onChange={e => setSlug(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
