import { useState, useEffect, useRef } from 'react'
import { request } from '../api/client'

export function usePolling<T>(url: string | null, intervalMs: number, enabled: boolean) {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<string | null>(null)
  const timerRef = useRef<number | null>(null)

  useEffect(() => {
    if (!enabled || !url) return
    const fetchData = async () => {
      try {
        const result = await request<T>('GET', url)
        setData(result)
        setError(null)
      } catch (e: any) { setError(e.message) }
    }
    fetchData()
    timerRef.current = window.setInterval(fetchData, intervalMs)
    return () => { if (timerRef.current) clearInterval(timerRef.current) }
  }, [url, intervalMs, enabled])

  return { data, error }
}
