import { useState } from 'react'
import { request } from '../api/client'

interface Subject { id: number; title: string; slug: string; color: string }
interface Session { id: number; subjectId: number; title: string; description: string; orderIndex: number }
interface Lesson { id: number; subjectSessionId: number; title: string; contentMd: string; orderIndex: number }
interface ProblemItem { id: number; slug: string; title: string; difficulty: string; orderIndex: number; acceptanceRate: number }

export default function Curriculum() {
  const [subjects, setSubjects] = useState<Subject[]>([])
  const [sessions, setSessions] = useState<Session[]>([])
  const [lessons, setLessons] = useState<Lesson[]>([])
  const [problems, setProblems] = useState<ProblemItem[]>([])
  const [error, setError] = useState('')

  const loadSubjects = async () => {
    try { const d = await request<any>('GET', '/subjects'); setSubjects(d.subjects || []); setSessions([]); setLessons([]); setProblems([]) }
    catch (e: any) { setError(e.message) }
  }

  const loadSessions = async (slug: string) => {
    try { const d = await request<any>('GET', `/subjects/${slug}/sessions`); setSessions(d.sessions || []); setLessons([]); setProblems([]) }
    catch (e: any) { setError(e.message) }
  }

  const loadLessons = async (sessionId: number) => {
    try { const d = await request<any>('GET', `/sessions/${sessionId}/lessons`); setLessons(d.lessons || []); setProblems([]) }
    catch (e: any) { setError(e.message) }
  }

  const loadProblems = async (lessonId: number) => {
    try { const d = await request<any>('GET', `/lessons/${lessonId}/problems`); setProblems(d.problems || []) }
    catch (e: any) { setError(e.message) }
  }

  return (
    <div>
      <h2>Curriculum</h2>
      {error && <p className="error">{error}</p>}
      <button onClick={loadSubjects}>Load Subjects</button>
      <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
        <div style={{ flex: 1 }}>
          <h4>Subjects</h4>
          {subjects.map(s => <div key={s.id} className="card" style={{ cursor: 'pointer' }} onClick={() => loadSessions(s.slug)}>{s.title} <small>({s.slug})</small></div>)}
        </div>
        <div style={{ flex: 1 }}>
          <h4>Sessions</h4>
          {sessions.map(s => <div key={s.id} className="card" style={{ cursor: 'pointer' }} onClick={() => loadLessons(s.id)}>{s.title} <small>order: {s.orderIndex}</small></div>)}
        </div>
        <div style={{ flex: 1 }}>
          <h4>Lessons</h4>
          {lessons.map(l => <div key={l.id} className="card" style={{ cursor: 'pointer' }} onClick={() => loadProblems(l.id)}>{l.title}</div>)}
        </div>
        <div style={{ flex: 1 }}>
          <h4>Problems</h4>
          {problems.map(p => <div key={p.id} className="card">{p.title} <span className="status-badge">{p.difficulty}</span></div>)}
        </div>
      </div>
    </div>
  )
}
