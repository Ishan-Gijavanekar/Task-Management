import { afterEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { ApiError } from './errors'
import { tokenStore } from './token-store'

afterEach(() => {
  vi.restoreAllMocks()
  tokenStore.set(null)
})

describe('api client', () => {
  it('maps the public error envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: {
              code: 'INVALID_CREDENTIALS',
              message: 'Email or password is incorrect',
              requestId: 'req-7',
            },
          }),
          { status: 401, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )

    await expect(
      api.login({ email: 'person@example.com', password: 'wrong-pass' }),
    ).rejects.toMatchObject<Partial<ApiError>>({
      status: 401,
      code: 'INVALID_CREDENTIALS',
      requestId: 'req-7',
    })
  })

  it('keeps access tokens in memory and sends cookies for refresh', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: 'short-lived-token',
          expires_in: 900,
          token_type: 'Bearer',
          user: {
            id: 'u1',
            name: 'Asha',
            email: 'asha@example.com',
            role: 'user',
            status: 'active',
            version: 1,
            createdAt: '',
            updatedAt: '',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)

    await api.refresh()

    expect(tokenStore.get()).toBe('short-lived-token')
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/auth/refresh',
      expect.objectContaining({ credentials: 'include', method: 'POST' }),
    )
    expect(localStorage.length).toBe(0)
  })

  it('normalizes empty Go collections before pages render them', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ projects: null }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ members: null }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ tasks: null }), { status: 200 }),
      )
    vi.stubGlobal('fetch', fetchMock)

    await expect(api.listProjects('workspace-1')).resolves.toEqual({
      projects: [],
    })
    await expect(api.listMembers('workspace-1')).resolves.toEqual({
      members: [],
    })
    await expect(
      api.listTasks('workspace-1', new URLSearchParams()),
    ).resolves.toEqual({ tasks: [] })
  })

  it('creates and accepts workspace invitations', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({ invitation: {}, token: 'invite-token' }),
          {
            status: 201,
          },
        ),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ workspaceId: 'workspace-1' }), {
          status: 200,
        }),
      )
    vi.stubGlobal('fetch', fetchMock)

    await api.inviteMember('workspace-1', {
      email: 'person@example.com',
      role: 'member',
    })
    await api.acceptInvitation('invite-token')

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      '/api/v1/workspaces/workspace-1/invitations',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'person@example.com', role: 'member' }),
      }),
    )
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/api/v1/invitations/accept',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ token: 'invite-token' }),
      }),
    )
  })
})
