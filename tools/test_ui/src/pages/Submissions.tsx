import { useState } from 'react'
import { request } from '../api/client'
import { usePolling } from '../hooks/usePolling'
import CodeEditor from '../components/CodeEditor'
import StatusBadge from '../components/StatusBadge'

export default function Submissions() {
  const [problemId, setProblemId] = useState('1')
  const [problemSlug, setProblemSlug] = useState('sum-two-numbers')
  const [languageId, setLanguageId] = useState('1')
  const [code, setCode] = useState("const fs = require('fs');\nconst input = fs.readFileSync(0, 'utf8').trim().split(' ').map(Number);\nconsole.log(input[0] + input[1]);")
  const [submission, setSubmission] = useState<any>(null)
  const [error, setError] = useState('')
  const [pollUrl, setPollUrl] = useState<string | null>(null)
  const [list, setList] = useState<any[]>([])

  const submit = async () => {
    setError('')
    try {
      const d = await request<any>('POST', '/submit', { problemId: +problemId, languageId: +languageId, code })
      setSubmission(d.submission || d)
      setPollUrl(`/submissions/${d.submission.id}`)
    } catch (e: any) { setError(e.message) }
  }

  const pollResult = usePolling<any>(pollUrl, 2000, !!pollUrl)
  const finalStatus = pollResult.data?.submission?.status || pollResult.data?.status

  if (finalStatus && !['Pending', 'Running'].includes(finalStatus)) {
    if (pollUrl) setTimeout(() => setPollUrl(null), 0)
  }

  const loadList = async () => {
    try { const d = await request<any>('GET', '/me/submissions'); setList(d.submissions || []) }
    catch (e: any) { setError(e.message) }
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
      <h2>Submissions</h2>
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
      <button onClick={submit}>Submit</button>
      <button onClick={loadList} style={{ marginLeft: '0.5rem' }}>List My Submissions</button>
      {error && <p className="error">{error}</p>}
      {submission && <pre>{JSON.stringify(submission, null, 2)}</pre>}
      {pollResult.data && (
        <div className="card">
          <h4>Poll Result</h4>
          <StatusBadge status={finalStatus} />
          {pollResult.data?.submission && (
            <div style={{ marginTop: '0.75rem' }}>
              <p>Passed: {pollResult.data.submission.passCount} / {pollResult.data.submission.totalTestcases}</p>
              {pollResult.data.submission.runtimeMs && <p>Runtime: {pollResult.data.submission.runtimeMs}ms</p>}
              {pollResult.data.submission.repoPath && (
                <div style={{ marginTop: '0.5rem' }}>
                  <strong>Commit:</strong>
                  <pre style={{ margin: '0.25rem 0' }}>{pollResult.data.submission.repoPath}</pre>
                  {pollResult.data.submission.commitSha && <p>SHA: <code>{pollResult.data.submission.commitSha.substring(0, 7)}</code></p>}
                  {pollResult.data.submission.commitUrl && (
                    <p>🔗 <a href={pollResult.data.submission.commitUrl} target="_blank" rel="noopener">View on GitHub</a></p>
                  )}
                </div>
              )}
            </div>
          )}
          <pre style={{ marginTop: '0.75rem' }}>{JSON.stringify(pollResult.data, null, 2)}</pre>
        </div>
      )}
      {list.length > 0 && (
        <div className="card">
          <h4>My Submissions ({list.length})</h4>
          {list.slice(0, 10).map((s: any, i: number) => (
            <div key={s.id || i} style={{ padding: '0.5rem 0', borderBottom: i < list.slice(0, 10).length - 1 ? '1px solid #eee' : 'none' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
                <StatusBadge status={s.status} />
                <strong>{s.problem?.title || `Problem #${s.problemId}`}</strong>
              </div>
              <p style={{ fontSize: '0.85rem', color: '#666' }}>
                Passed: {s.passCount} / {s.totalTestcases}
                {s.runtimeMs && <> · Runtime: {s.runtimeMs}ms</>}
              </p>
              {s.commitUrl && (
                <p style={{ fontSize: '0.85rem' }}>
                  🔗 <a href={s.commitUrl} target="_blank" rel="noopener">View commit on GitHub</a>
                  {s.commitSha && <> (<code>{s.commitSha.substring(0, 7)}</code>)</>}
                </p>
              )}
              {s.repoPath && <p style={{ fontSize: '0.8rem', color: '#999' }}>{s.repoPath}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
