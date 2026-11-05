import React, { createContext, useContext, useState } from 'react'
import { getBaseURL, setBaseURL } from '../api/client'

interface ConfigCtx {
  baseURL: string
  setURL: (url: string) => void
}

const Ctx = createContext<ConfigCtx>({ baseURL: 'http://localhost:8080', setURL: () => {} })
export const useConfig = () => useContext(Ctx)

export function ConfigProvider({ children }: { children: React.ReactNode }) {
  const [baseURL, setBase] = useState(getBaseURL())
  const setURL = (url: string) => { setBaseURL(url); setBase(url) }
  return <Ctx.Provider value={{ baseURL, setURL }}>{children}</Ctx.Provider>
}
