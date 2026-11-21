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
})
