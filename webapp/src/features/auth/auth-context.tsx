/* eslint-disable react-refresh/only-export-components -- Provider and hook form one auth module. */
import { useQueryClient } from '@tanstack/react-query'
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { api, setSessionExpiredHandler } from '../../api/client'
import { tokenStore } from '../../api/token-store'
import type { LoginRequest, RegisterRequest, User } from '../../api/types'

type AuthStatus = 'loading' | 'authenticated' | 'anonymous'

interface AuthContextValue {
  status: AuthStatus
  user: User | null
  sessionExpired: boolean
  login(input: LoginRequest): Promise<void>
  register(input: RegisterRequest): Promise<void>
  logout(allSessions?: boolean): Promise<void>
  clearExpiredNotice(): void
}

const AuthContext = createContext<AuthContextValue | null>(null)
const SESSION_EVENT = 'taskflow-session'
const sessionChannel =
  typeof BroadcastChannel === 'undefined'
    ? null
    : new BroadcastChannel(SESSION_EVENT)

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const [status, setStatus] = useState<AuthStatus>('loading')
  const [user, setUser] = useState<User | null>(null)
  const [sessionExpired, setSessionExpired] = useState(false)

  const clearSession = useCallback(
    (expired = false) => {
      tokenStore.set(null)
      setUser(null)
      setStatus('anonymous')
      setSessionExpired(expired)
      queryClient.clear()
    },
    [queryClient],
  )

  const acceptSession = useCallback(
    (session: { access_token: string; user: User }) => {
      tokenStore.set(session.access_token)
      setUser(session.user)
      setStatus('authenticated')
      setSessionExpired(false)
    },
    [],
  )

  useEffect(() => {
    let active = true
    setSessionExpiredHandler(() => active && clearSession(true))
    api
      .refresh()
      .then((session) => active && acceptSession(session))
      .catch(() => active && clearSession())
    return () => {
      active = false
    }
  }, [acceptSession, clearSession])

  useEffect(() => {
    const onMessage = (event: MessageEvent<string>) => {
      if (event.data === 'logout') clearSession()
      if (event.data === 'login') {
        setStatus('loading')
        api
          .refresh()
          .then(acceptSession)
          .catch(() => clearSession())
      }
    }
    sessionChannel?.addEventListener('message', onMessage)
    return () => sessionChannel?.removeEventListener('message', onMessage)
  }, [acceptSession, clearSession])

  const value = useMemo<AuthContextValue>(
    () => ({
      status,
      user,
      sessionExpired,
      login: async (input) => {
        acceptSession(await api.login(input))
        sessionChannel?.postMessage('login')
      },
      register: async (input) => {
        acceptSession(await api.register(input))
        sessionChannel?.postMessage('login')
      },
      logout: async (allSessions = false) => {
        try {
          await api.logout(allSessions)
        } finally {
          clearSession()
          sessionChannel?.postMessage('logout')
        }
      },
      clearExpiredNotice: () => setSessionExpired(false),
    }),
    [acceptSession, clearSession, sessionExpired, status, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth must be used inside AuthProvider')
  return value
}
