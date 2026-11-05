const STORAGE_KEY_BASE = 'innogen_base_url'
const STORAGE_KEY_TOKEN = 'innogen_token'

export function getBaseURL(): string {
  return localStorage.getItem(STORAGE_KEY_BASE) || 'http://localhost:8080'
}

export function setBaseURL(url: string): void {
  localStorage.setItem(STORAGE_KEY_BASE, url)
}

export function getToken(): string | null {
  return localStorage.getItem(STORAGE_KEY_TOKEN)
}

export function setToken(t: string | null): void {
  if (t) localStorage.setItem(STORAGE_KEY_TOKEN, t)
  else localStorage.removeItem(STORAGE_KEY_TOKEN)
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

export async function request<T = any>(method: string, path: string, body?: any): Promise<T> {
  const base = getBaseURL()
  const url = `${base}${path}`
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  const json = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, json.message || json.error || 'Unknown error')
  }

  if (res.status === 204) return {} as T
  return json.data !== undefined ? json.data : json
}
