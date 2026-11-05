import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { request } from '../api/client'

export default function Problem() {
  const { slug } = useParams()
  const [problem, setProblem] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try { const d = await request<any>('GET', `/problems/${slug}`); setProblem(d.problem || d) }
    catch (e: any) { alert(e.message) }
    finally { setLoading(false) }
  }

  return (
    <div>
      <h2>Problem: {slug}</h2>
      <button onClick={load} disabled={loading}>Load</button>
      {problem && <pre>{JSON.stringify(problem, null, 2)}</pre>}
    </div>
  )
}
