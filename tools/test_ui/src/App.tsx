import { Routes, Route } from 'react-router-dom'
import Navbar from './components/Navbar'
import Dashboard from './pages/Dashboard'
import Login from './pages/Login'
import Curriculum from './pages/Curriculum'
import Problem from './pages/Problem'
import Runner from './pages/Runner'
import Submissions from './pages/Submissions'
import RepoNotes from './pages/RepoNotes'
import GithubConnect from './pages/GithubConnect'
import AdminLayout from './pages/admin/AdminLayout'
import AdminLanguages from './pages/admin/Languages'
import AdminSubjects from './pages/admin/Subjects'
import AdminSessions from './pages/admin/Sessions'
import AdminLessons from './pages/admin/Lessons'
import AdminProblems from './pages/admin/Problems'
import AdminTestCases from './pages/admin/TestCases'
import AdminTags from './pages/admin/Tags'
import { AuthProvider } from './hooks/useAuth'
import { ConfigProvider } from './hooks/useConfig'

export default function App() {
  return (
    <ConfigProvider>
      <AuthProvider>
        <div className="app">
          <Navbar />
          <main className="content">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/login" element={<Login />} />
              <Route path="/curriculum" element={<Curriculum />} />
              <Route path="/problem/:slug" element={<Problem />} />
              <Route path="/runner" element={<Runner />} />
              <Route path="/submissions" element={<Submissions />} />
              <Route path="/repo-notes" element={<RepoNotes />} />
              <Route path="/github/callback" element={<GithubConnect />} />
              <Route path="/github/connect" element={<GithubConnect />} />
              <Route path="/admin" element={<AdminLayout />}>
                <Route index element={<AdminLanguages />} />
                <Route path="languages" element={<AdminLanguages />} />
                <Route path="subjects" element={<AdminSubjects />} />
                <Route path="sessions" element={<AdminSessions />} />
                <Route path="lessons" element={<AdminLessons />} />
                <Route path="problems" element={<AdminProblems />} />
                <Route path="test-cases" element={<AdminTestCases />} />
                <Route path="tags" element={<AdminTags />} />
              </Route>
            </Routes>
          </main>
        </div>
      </AuthProvider>
    </ConfigProvider>
  )
}
