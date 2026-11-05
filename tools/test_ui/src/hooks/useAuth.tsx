import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { request, setToken, getToken, ApiError } from '../api/client'

interface User { id: number; email: string; username: string; fullName: string; role: string }

interface AuthCtx {
  token: string | null
  user: User | null
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  loading: boolean
}

const Ctx = createContext<AuthCtx>({ token: null, user: null, login: async () => {}, logout: () => {}, loading: false })
export const useAuth = () => useContext(Ctx)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTokenState] = useState<string | null>(getToken())
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchMe = useCallback(async () => {
    try {
      const data = await request<any>('GET', '/auth/me')
      setUser(data.user || data)
    } catch { setUser(null) }
  }, [])

  useEffect(() => { if (token) fetchMe() }, [token])

  const login = async (email: string, password: string) => {
    setLoading(true)
    try {
      const data = await request<any>('POST', '/auth/login', { email, password })
      const t = data.accessToken
      setToken(t)
      setTokenState(t)
      fetchMe()
    } finally { setLoading(false) }
  }

  const logout = () => { setToken(null); setTokenState(null); setUser(null) }

  return <Ctx.Provider value={{ token, user, login, logout, loading }}>{children}</Ctx.Provider>
}
