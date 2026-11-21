import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import axe from 'axe-core'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AuthProvider } from './auth-context'
import { AuthLayout } from './auth-layout'
import { LoginPage } from './auth-pages'
import { AnonymousRoute } from './route-guards'

function renderLogin() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={['/login']}>
        <AuthProvider>
          <Routes>
            <Route element={<AnonymousRoute />}>
              <Route element={<AuthLayout />}>
                <Route path="login" element={<LoginPage />} />
              </Route>
            </Route>
            <Route path="/" element={<h1>Overview</h1>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue(new Response('', { status: 401 })),
  )
})

describe('login page', () => {
  it('is keyboard-operable and has no automatic accessibility violations', async () => {
    renderLogin()
    const button = await screen.findByRole('button', { name: 'Sign in' })
    screen.getByLabelText('Work email').focus()
    await userEvent.tab()
    expect(screen.getByRole('link', { name: 'Forgot password?' })).toHaveFocus()
    await userEvent.tab()
    expect(screen.getByLabelText('Password')).toHaveFocus()
    const audit = await axe.run(document.body, {
      rules: { 'color-contrast': { enabled: false } },
    })
    expect(audit.violations).toEqual([])
    expect(button).toBeEnabled()
  })

  it('shows a safe API error without losing entered values', async () => {
    renderLogin()
    const email = await screen.findByLabelText('Work email')
    await userEvent.type(email, 'asha@example.com')
    await userEvent.type(screen.getByLabelText('Password'), 'incorrect')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent(
        'session has expired',
      ),
    )
    expect(email).toHaveValue('asha@example.com')
  })
})
