import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Route } from '@playwright/test'

const now = '2026-11-21T08:00:00.000Z'
const user = {
  id: 'user-e2e',
  name: 'Asha Rao',
  email: 'asha.e2e@example.com',
  role: 'user',
  status: 'active',
  version: 1,
  createdAt: now,
  updatedAt: now,
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

test('runs the primary workspace and task workflow', async ({
  page,
  context,
}) => {
  const workspace = {
    id: 'workspace-1',
    name: 'Product studio',
    ownerId: user.id,
    version: 1,
    createdAt: now,
    updatedAt: now,
  }
  const projects: Record<string, unknown>[] = []
  const tasks: Record<string, unknown>[] = []

  await context.route('**/api/v1/**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const path = url.pathname.replace('/api/v1', '')
    const method = request.method()

    if (path === '/auth/refresh') return json(route, session)
    if (path === '/workspaces' && method === 'POST')
      return json(route, workspace, 201)
    if (path === `/workspaces/${workspace.id}` && method === 'GET')
      return json(route, workspace)
    if (path === `/workspaces/${workspace.id}/members`)
      return json(route, {
        members: [
          {
            workspaceId: workspace.id,
            userId: user.id,
            role: 'owner',
            status: 'active',
          },
        ],
      })
    if (path === `/workspaces/${workspace.id}/projects` && method === 'GET')
      return json(route, { projects })
    if (path === `/workspaces/${workspace.id}/projects` && method === 'POST') {
      const input = request.postDataJSON()
      const project = {
        ...input,
        id: 'project-1',
        workspaceId: workspace.id,
        version: 1,
      }
      projects.push(project)
      return json(route, project, 201)
    }
    if (path === `/workspaces/${workspace.id}/tasks` && method === 'GET') {
      const projectId = url.searchParams.get('projectId')
      const query = url.searchParams.get('q')?.toLowerCase()
      const filtered = tasks.filter(
        (task) =>
          (!projectId || task.projectId === projectId) &&
          (!query || String(task.title).toLowerCase().includes(query)),
      )
      return json(route, { tasks: filtered })
    }
    if (path === `/workspaces/${workspace.id}/tasks` && method === 'POST') {
      expect(request.headers()['idempotency-key']).toBeTruthy()
      const input = request.postDataJSON()
      const task = {
        ...input,
        id: 'task-1',
        workspaceId: workspace.id,
        reporterId: user.id,
        status: 'backlog',
        position: 1024,
        version: 1,
        createdAt: now,
        updatedAt: now,
      }
      tasks.push(task)
      return json(route, task, 201)
    }
    if (path === `/workspaces/${workspace.id}/tasks/task-1`) {
      const index = tasks.findIndex((task) => task.id === 'task-1')
      if (method === 'GET') return json(route, tasks[index])
      if (method === 'PATCH') {
        expect(request.headers()['idempotency-key']).toBeTruthy()
        const input = request.postDataJSON()
        tasks[index] = {
          ...tasks[index],
          ...input,
          version: Number(tasks[index].version) + 1,
          updatedAt: now,
        }
        return json(route, tasks[index])
      }
    }
    return json(
      route,
      { error: { code: 'NOT_FOUND', message: `${method} ${path}` } },
      404,
    )
  })

  await page.goto('/workspaces')
  await page.getByLabel('Name', { exact: true }).fill(workspace.name)
  await page.getByRole('button', { name: 'Create workspace' }).click()
  await expect(
    page.getByText(workspace.name, { exact: true }).first(),
  ).toBeVisible()

  await page.goto('/projects')
  await page.getByRole('button', { name: /New project/ }).click()
  await page.getByLabel('Name', { exact: true }).fill('Launch board')
  await page.getByLabel('Key', { exact: true }).fill('WEB')
  await page.getByLabel('Description', { exact: true }).fill('Plan the release')
  await page
    .getByRole('dialog')
    .getByRole('button', { name: 'Create project' })
    .click()
  await expect(
    page.getByRole('heading', { name: 'Launch board' }),
  ).toBeVisible()
  await page.getByRole('link', { name: 'Open board' }).click()

  await page.getByRole('button', { name: /New task/ }).click()
  const taskDialog = page.getByRole('dialog')
  await taskDialog.getByLabel('Title').fill('Publish release notes')
  await taskDialog.getByLabel('Description').fill('Explain the new workflow')
  await taskDialog.getByLabel('Priority').selectOption('high')
  await taskDialog.getByLabel('Assignee').selectOption(user.id)
  await taskDialog.getByLabel('Checklist').fill('Draft notes\nReview copy')
  await taskDialog.getByRole('button', { name: 'Create task' }).click()

  await expect(page).toHaveURL(/task=task-1/)
  await expect(
    page.getByRole('heading', { name: 'Publish release notes' }),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Close task' }).click()
  await expect(page).not.toHaveURL(/task=/)

  await page.getByLabel('Move Publish release notes').selectOption('todo')
  await expect(page.getByLabel('Move Publish release notes')).toHaveValue(
    'todo',
  )

  await page.getByRole('searchbox', { name: 'Search tasks' }).fill('Publish')
  await page.getByLabel('Move Publish release notes').selectOption('todo')

  await expect(page.getByLabel('Move Publish release notes')).toHaveValue(
    'todo',
  )

  const taskSearch = page.getByRole('searchbox', {
    name: 'Search tasks',
  })

  await expect(taskSearch).toBeVisible()
  await taskSearch.fill('Publish')

  await expect(page).toHaveURL(/q=Publish/)

  await page.getByRole('button', { name: /Publish release notes/ }).click()
  await page.getByRole('button', { name: 'Edit task' }).click()

  await taskDialog
    .getByLabel('Description')
    .fill('Explain the complete workflow')

  await taskDialog.getByRole('button', { name: 'Save changes' }).click()

  await expect(
    taskDialog.getByText('Explain the complete workflow'),
  ).toBeVisible()

  await page.getByRole('button', { name: 'Close task' }).click()

  await expect(page).toHaveURL(/q=Publish/)

  const accessibility = await new AxeBuilder({ page }).analyze()
  expect(accessibility.violations).toEqual([])
})
