import { ApiError, mapApiError } from './errors'
import { tokenStore } from './token-store'
import type {
  AuthResponse,
  LoginRequest,
  MessageResponse,
  RegisterRequest,
  User,
  Workspace,
} from './types'

const API_ROOT = import.meta.env.VITE_API_URL ?? '/api/v1'
let refreshRequest: Promise<AuthResponse> | null = null
let onSessionExpired: (() => void) | undefined

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  skipRefresh?: boolean
}

export function setSessionExpiredHandler(handler: () => void) {
  onSessionExpired = handler
}

async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { body, skipRefresh, ...init } = options
  const headers = new Headers(init.headers)
  const token = tokenStore.get()
  if (body !== undefined) headers.set('Content-Type', 'application/json')
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const fetchOptions: RequestInit = {
    ...init,
    credentials: 'include',
    headers,
  }
  if (body !== undefined) fetchOptions.body = JSON.stringify(body)
  const response = await fetch(`${API_ROOT}${path}`, fetchOptions)

  if (response.status === 401 && !skipRefresh && path !== '/auth/refresh') {
    try {
      await refreshSession()
      return request<T>(path, { ...options, skipRefresh: true })
    } catch {
      tokenStore.set(null)
      onSessionExpired?.()
    }
  }

  if (!response.ok) throw await mapApiError(response)
  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

export function refreshSession() {
  refreshRequest ??= request<AuthResponse>('/auth/refresh', {
    method: 'POST',
    body: {},
    skipRefresh: true,
  })
    .then((session) => {
      tokenStore.set(session.access_token)
      return session
    })
    .finally(() => {
      refreshRequest = null
    })
  return refreshRequest
}

export const api = {
  register: (input: RegisterRequest, signal?: AbortSignal) =>
    request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: input,
      signal: signal ?? null,
      skipRefresh: true,
    }),
  login: (input: LoginRequest, signal?: AbortSignal) =>
    request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: input,
      signal: signal ?? null,
      skipRefresh: true,
    }),
  refresh: refreshSession,
  logout: (allSessions = false) =>
    request<void>('/auth/logout', {
      method: 'POST',
      body: { all_sessions: allSessions },
      skipRefresh: true,
    }),
  forgotPassword: (email: string, signal?: AbortSignal) =>
    request<MessageResponse>('/auth/forgot-password', {
      method: 'POST',
      body: { email },
      signal: signal ?? null,
      skipRefresh: true,
    }),
  resetPassword: (token: string, newPassword: string, signal?: AbortSignal) =>
    request<MessageResponse>('/auth/reset-password', {
      method: 'POST',
      body: { token, new_password: newPassword },
      signal: signal ?? null,
      skipRefresh: true,
    }),
  verifyEmail: (token: string, signal?: AbortSignal) =>
    request<MessageResponse>('/auth/verify-email', {
      method: 'POST',
      body: { token },
      signal: signal ?? null,
      skipRefresh: true,
    }),
  resendVerification: (email: string, signal?: AbortSignal) =>
    request<MessageResponse>('/auth/resend-verification', {
      method: 'POST',
      body: { email },
      signal: signal ?? null,
      skipRefresh: true,
    }),
  me: (signal?: AbortSignal) => request<User>('/me', { signal: signal ?? null }),
  getWorkspace: (id: string, signal?: AbortSignal) =>
    request<Workspace>(`/workspaces/${id}`, { signal: signal ?? null }),
}

export function shouldRetry(failureCount: number, error: unknown) {
  if (failureCount >= 2) return false
  if (error instanceof ApiError)
    return error.status === 429 || error.status >= 500
  return error instanceof TypeError
}
