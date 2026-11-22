import { ApiError, mapApiError } from './errors'
import { tokenStore } from './token-store'
import type {
  AuthResponse,
  ActivityPage,
  CommentPage,
  CommentRecord,
  Label,
  LabelList,
  CreateProjectRequest,
  CreateWorkspaceRequest,
  InvitationResult,
  InviteRequest,
  LoginRequest,
  MemberList,
  MessageResponse,
  NotificationList,
  NotificationPreferences,
  Project,
  ProjectList,
  RegisterRequest,
  TaskPage,
  TaskRecord,
  UpdateProjectRequest,
  UpdateWorkspaceRequest,
  User,
  Workspace,
  Membership,
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
  me: (signal?: AbortSignal) =>
    request<User>('/me', { signal: signal ?? null }),
  getWorkspace: (id: string, signal?: AbortSignal) =>
    request<Workspace>(`/workspaces/${id}`, { signal: signal ?? null }),
  createWorkspace: (input: CreateWorkspaceRequest, signal?: AbortSignal) =>
    request<Workspace>('/workspaces', {
      method: 'POST',
      body: input,
      signal: signal ?? null,
    }),
  updateWorkspace: (
    id: string,
    input: UpdateWorkspaceRequest,
    signal?: AbortSignal,
  ) =>
    request<Workspace>(`/workspaces/${id}`, {
      method: 'PATCH',
      body: input,
      signal: signal ?? null,
    }),
  listProjects: (workspaceId: string, signal?: AbortSignal) =>
    request<ProjectList>(`/workspaces/${workspaceId}/projects`, {
      signal: signal ?? null,
    }).then((result) => ({ ...result, projects: result.projects ?? [] })),
  getProject: (workspaceId: string, projectId: string, signal?: AbortSignal) =>
    request<Project>(`/workspaces/${workspaceId}/projects/${projectId}`, {
      signal: signal ?? null,
    }),
  createProject: (
    workspaceId: string,
    input: CreateProjectRequest,
    signal?: AbortSignal,
  ) =>
    request<Project>(`/workspaces/${workspaceId}/projects`, {
      method: 'POST',
      body: input,
      signal: signal ?? null,
    }),
  updateProject: (
    workspaceId: string,
    projectId: string,
    input: UpdateProjectRequest,
    signal?: AbortSignal,
  ) =>
    request<Project>(`/workspaces/${workspaceId}/projects/${projectId}`, {
      method: 'PATCH',
      body: input,
      signal: signal ?? null,
    }),
  listMembers: (workspaceId: string, signal?: AbortSignal) =>
    request<MemberList>(`/workspaces/${workspaceId}/members`, {
      signal: signal ?? null,
    }).then((result) => ({ ...result, members: result.members ?? [] })),
  inviteMember: (workspaceId: string, input: InviteRequest) =>
    request<InvitationResult>(`/workspaces/${workspaceId}/invitations`, {
      method: 'POST',
      body: input,
    }),
  acceptInvitation: (token: string) =>
    request<Membership>('/invitations/accept', {
      method: 'POST',
      body: { token },
    }),
  listLabels: (workspaceId: string, signal?: AbortSignal) =>
    request<LabelList>(`/workspaces/${workspaceId}/labels`, {
      signal: signal ?? null,
    }).then((result) => ({ ...result, labels: result.labels ?? [] })),
  createLabel: (workspaceId: string, name: string, color: string) =>
    request<Label>(`/workspaces/${workspaceId}/labels`, {
      method: 'POST',
      body: { name, color },
    }),
  deleteLabel: (workspaceId: string, labelId: string) =>
    request<void>(`/workspaces/${workspaceId}/labels/${labelId}`, {
      method: 'DELETE',
    }),
  listTasks: (
    workspaceId: string,
    search: URLSearchParams,
    signal?: AbortSignal,
  ) =>
    request<TaskPage>(`/workspaces/${workspaceId}/tasks?${search.toString()}`, {
      signal: signal ?? null,
    }).then((result) => ({ ...result, tasks: result.tasks ?? [] })),
  getTask: (
    workspaceId: string,
    taskId: string,
    includeDeleted = false,
    signal?: AbortSignal,
  ) =>
    request<TaskRecord>(
      `/workspaces/${workspaceId}/tasks/${taskId}?includeDeleted=${includeDeleted}`,
      { signal: signal ?? null },
    ),
  createTask: (
    workspaceId: string,
    input: Record<string, unknown>,
    idempotencyKey: string,
  ) =>
    request<TaskRecord>(`/workspaces/${workspaceId}/tasks`, {
      method: 'POST',
      body: input,
      headers: { 'Idempotency-Key': idempotencyKey },
    }),
  updateTask: (
    workspaceId: string,
    taskId: string,
    input: Record<string, unknown>,
    idempotencyKey: string,
  ) =>
    request<TaskRecord>(`/workspaces/${workspaceId}/tasks/${taskId}`, {
      method: 'PATCH',
      body: input,
      headers: { 'Idempotency-Key': idempotencyKey },
    }),
  archiveTask: (
    workspaceId: string,
    taskId: string,
    version: number,
    idempotencyKey: string,
  ) =>
    request<void>(
      `/workspaces/${workspaceId}/tasks/${taskId}?version=${version}`,
      {
        method: 'DELETE',
        headers: { 'Idempotency-Key': idempotencyKey },
      },
    ),
  restoreTask: (
    workspaceId: string,
    taskId: string,
    version: number,
    idempotencyKey: string,
  ) =>
    request<TaskRecord>(`/workspaces/${workspaceId}/tasks/${taskId}/restore`, {
      method: 'POST',
      body: { version },
      headers: { 'Idempotency-Key': idempotencyKey },
    }),
  listComments: (workspaceId: string, taskId: string, signal?: AbortSignal) =>
    request<CommentPage>(
      `/workspaces/${workspaceId}/tasks/${taskId}/comments?limit=100`,
      { signal: signal ?? null },
    ).then((result) => ({ ...result, comments: result.comments ?? [] })),
  createComment: (
    workspaceId: string,
    taskId: string,
    body: string,
    mentionIds: string[],
  ) =>
    request<CommentRecord>(
      `/workspaces/${workspaceId}/tasks/${taskId}/comments`,
      { method: 'POST', body: { body, mentionIds } },
    ),
  listActivities: (workspaceId: string, taskId: string, signal?: AbortSignal) =>
    request<ActivityPage>(
      `/workspaces/${workspaceId}/tasks/${taskId}/activities?limit=100`,
      { signal: signal ?? null },
    ).then((result) => ({ ...result, activities: result.activities ?? [] })),
  listNotifications: (unread = false, signal?: AbortSignal) =>
    request<NotificationList>(`/notifications?unread=${unread}&limit=100`, {
      signal: signal ?? null,
    }).then((result) => ({
      ...result,
      notifications: result.notifications ?? [],
    })),
  markNotificationsRead: (notificationId?: string) =>
    request<void>(
      notificationId
        ? `/notifications/${notificationId}/read`
        : '/notifications/read',
      { method: 'POST', body: {} },
    ),
  getNotificationPreferences: (signal?: AbortSignal) =>
    request<NotificationPreferences>('/notification-preferences', {
      signal: signal ?? null,
    }),
  updateNotificationPreferences: (preferences: NotificationPreferences) =>
    request<NotificationPreferences>('/notification-preferences', {
      method: 'PUT',
      body: preferences,
    }),
}

export function shouldRetry(failureCount: number, error: unknown) {
  if (failureCount >= 2) return false
  if (error instanceof ApiError)
    return error.status === 429 || error.status >= 500
  return error instanceof TypeError
}
