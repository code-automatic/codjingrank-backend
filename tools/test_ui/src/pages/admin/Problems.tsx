import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminProblems() {
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [difficulty, setDifficulty] = useState('')
  const [problemMd, setProblemMd] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/problems', { slug, title, difficulty, problemMd })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/problems'); setList(d.problems || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Problems</h3>
      <div className="form-group"><label>Slug</label><input value={slug} onChange={e => setSlug(e.target.value)} /></div>
      <div className="form-group"><label>Title</label><input value={title} onChange={e => setTitle(e.target.value)} /></div>
      <div className="form-group"><label>Difficulty</label><input value={difficulty} onChange={e => setDifficulty(e.target.value)} /></div>
      <div className="form-group"><label>Problem MD</label><textarea value={problemMd} onChange={e => setProblemMd(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
