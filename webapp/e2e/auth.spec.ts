import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page, type Route } from '@playwright/test'

const user = {
  id: 'user-e2e',
  name: 'Asha Rao',
  email: 'asha.e2e@example.com',
  role: 'user',
  status: 'active',
  version: 1,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}
const session = {
  access_token: 'e2e-access-token',
  expires_in: 900,
  token_type: 'Bearer',
  user,
}

async function json(route: Route, body: unknown, status = 200) {
  await route.fulfill({
    status,
    contentType: 'application/json',
    body: JSON.stringify(body),
  })
}

async function anonymous(page: Page) {
  await page.route('**/api/v1/auth/refresh', (route) =>
    json(
      route,
      { error: { code: 'UNAUTHENTICATED', message: 'Sign in' } },
      401,
    ),
  )
}

test('registers with isolated data', async ({ page }) => {
  await anonymous(page)
  await page.route('**/api/v1/auth/register', (route) =>
    json(route, session, 201),
  )
  await page.goto('/register')
  await page.getByLabel('Full name').fill(user.name)
  await page.getByLabel('Work email').fill(user.email)
  await page.getByLabel('Password').fill('StrongPass123')
  await page.getByRole('button', { name: 'Create account' }).click()
  await expect(
    page.getByRole('heading', { name: 'Create or open a workspace' }),
  ).toBeVisible()
})

test('signs in and restores the requested route', async ({ page }) => {
  await anonymous(page)
  await page.route('**/api/v1/auth/login', (route) => json(route, session))
  await page.goto('/login?returnTo=%2Ftasks')
  await page.getByLabel('Work email').fill(user.email)
  await page.getByLabel('Password').fill('StrongPass123')
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page).toHaveURL(/\/tasks$/)
  await expect(
    page.getByRole('heading', { name: 'Create or open a workspace' }),
  ).toBeVisible()
})

test('restores a refresh-cookie session and logs out across tabs', async ({
  page,
  context,
}) => {
  await context.route('**/api/v1/auth/refresh', (route) => json(route, session))
  await context.route('**/api/v1/auth/logout', (route) =>
    route.fulfill({ status: 204 }),
  )
  await page.goto('/')
  await expect(
    page.getByRole('heading', { name: 'Create or open a workspace' }),
  ).toBeVisible()
  const menuButton = page.getByRole('button', { name: 'More' })
  if (await menuButton.isVisible()) await menuButton.click()
  const secondPage = await context.newPage()
  await secondPage.goto('/tasks')
  await expect(
    secondPage.getByRole('heading', { name: 'Create or open a workspace' }),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Sign out' }).click()
  await expect(page).toHaveURL(/\/login/)
  await expect(secondPage).toHaveURL(/\/login/)
})

test('resets a password through a secure token', async ({ page }) => {
  await anonymous(page)
  await page.route('**/api/v1/auth/forgot-password', (route) =>
    json(route, { message: 'reset instructions sent' }, 202),
  )
  await page.route('**/api/v1/auth/reset-password', (route) =>
    json(route, { message: 'password reset successfully' }),
  )
  await page.goto('/forgot-password')
  await page.getByLabel('Work email').fill(user.email)
  await page.getByRole('button', { name: 'Send reset link' }).click()
  await expect(
    page.getByRole('heading', { name: 'Check your inbox' }),
  ).toBeVisible()
  await page.goto('/reset-password?token=isolated-reset-token')
  await page
    .getByLabel('New password', { exact: true })
    .fill('NewStrongPass123')
  await page
    .getByLabel('Confirm password', { exact: true })
    .fill('NewStrongPass123')
  await page.getByRole('button', { name: 'Update password' }).click()
  await expect(page).toHaveURL(/\/login\?reset=success/)
})

test('login page passes an automated accessibility scan', async ({ page }) => {
  await anonymous(page)
  await page.goto('/login')
  const results = await new AxeBuilder({ page }).analyze()
  expect(results.violations).toEqual([])
})
