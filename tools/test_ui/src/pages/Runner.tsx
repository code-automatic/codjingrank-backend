import { useState } from 'react'
import { request } from '../api/client'
import CodeEditor from '../components/CodeEditor'

export default function Runner() {
  const [problemId, setProblemId] = useState('1')
  const [problemSlug, setProblemSlug] = useState('sum-two-numbers')
  const [languageId, setLanguageId] = useState('1')
  const [code, setCode] = useState("const fs = require('fs');\nconst input = fs.readFileSync(0, 'utf8').trim().split(' ').map(Number);\nconsole.log(input[0] + input[1]);")
  const [result, setResult] = useState<any>(null)
  const [error, setError] = useState('')

  const run = async () => {
    setError('')
    try {
      const d = await request<any>('POST', '/run', { problemId: +problemId, languageId: +languageId, code })
      setResult(d)
    } catch (e: any) { setError(e.message) }
  }

  const loadInitialCode = async () => {
    setError('')
    try {
      const d = await request<any>('GET', `/problems/${problemSlug}`)
      if (d.problem?.initialCode) {
        setCode(d.problem.initialCode)
      } else {
        setError("No initialCode found for this problem")
      }
    } catch (e: any) { setError(e.message) }
  }

  return (
    <div>
      <h2>Runner</h2>
      <div className="form-group"><label>Problem ID</label><input value={problemId} onChange={e => setProblemId(e.target.value)} /></div>
      <div className="form-group">
        <label>Problem Slug</label>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <input value={problemSlug} onChange={e => setProblemSlug(e.target.value)} />
          <button onClick={loadInitialCode}>Load initialCode</button>
        </div>
      </div>
      <div className="form-group"><label>Language ID</label><input value={languageId} onChange={e => setLanguageId(e.target.value)} /></div>
      <CodeEditor value={code} onChange={setCode} label="Code" />
      <button onClick={run} style={{ marginBottom: '1rem' }}>Run</button>
      {error && <p className="error">{error}</p>}
      {result && <pre>{JSON.stringify(result, null, 2)}</pre>}
    </div>
  )
}
