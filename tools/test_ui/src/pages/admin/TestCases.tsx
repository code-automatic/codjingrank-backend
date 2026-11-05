import { useState } from 'react'
import { request } from '../../api/client'
export default function AdminTestCases() {
  const [problemId, setProblemId] = useState('')
  const [visibility, setVisibility] = useState('public')
  const [inputData, setInputData] = useState('')
  const [expectedOutput, setExpectedOutput] = useState('')
  const [orderIndex, setOrderIndex] = useState('')
  const [list, setList] = useState<any[]>([])
  const [msg, setMsg] = useState('')

  const create = async () => {
    try {
      await request('POST', '/admin/test-cases', {
        problemId: +problemId,
        visibility,
        inputData,
        expectedOutput,
        orderIndex: orderIndex ? +orderIndex : undefined,
      })
      setMsg('Created')
    } catch (e: any) { setMsg(e.message) }
  }
  const load = async () => {
    try { const d = await request<any>('GET', '/admin/test-cases'); setList(d.testCases || []) }
    catch (e: any) { setMsg(e.message) }
  }

  return (
    <div>
      <h3>Test Cases</h3>
      <div className="form-group"><label>Problem ID</label><input value={problemId} onChange={e => setProblemId(e.target.value)} /></div>
      <div className="form-group"><label>Visibility</label>
        <select value={visibility} onChange={e => setVisibility(e.target.value)}>
          <option value="public">public</option>
          <option value="private">private</option>
        </select>
      </div>
      <div className="form-group"><label>Input Data</label><textarea value={inputData} onChange={e => setInputData(e.target.value)} /></div>
      <div className="form-group"><label>Expected Output</label><textarea value={expectedOutput} onChange={e => setExpectedOutput(e.target.value)} /></div>
      <div className="form-group"><label>Order Index</label><input value={orderIndex} onChange={e => setOrderIndex(e.target.value)} /></div>
      <button onClick={create}>Create</button>
      <button onClick={load} style={{ marginLeft: '0.5rem' }}>List</button>
      {msg && <p className="error">{msg}</p>}
      {list.length > 0 && <pre>{JSON.stringify(list, null, 2)}</pre>}
    </div>
  )
}
