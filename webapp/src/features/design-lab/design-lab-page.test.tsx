import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'
import { DesignLabPage } from './design-lab-page'

function showLab() {
  return render(
    <MemoryRouter>
      <DesignLabPage />
    </MemoryRouter>,
  )
}

beforeEach(() => localStorage.clear())

describe('Interactive UI lab', () => {
  it('searches across task title, project and assignee', async () => {
    const user = userEvent.setup()
    showLab()
    await screen.findByRole(
      'button',
      { name: /^Review workspace invitation permissions/ },
      { timeout: 2500 },
    )
    const search = screen.getByRole('searchbox')
    await user.type(search, 'Maya Iyer')
    expect(
      screen.getByRole('button', {
        name: /^Investigate duplicate task events/,
      }),
    ).toBeVisible()
    expect(
      screen.queryByRole('button', {
        name: /^Review workspace invitation permissions/,
      }),
    ).not.toBeInTheDocument()
    await user.clear(search)
    expect(
      screen.getByRole('button', {
        name: /^Review workspace invitation permissions/,
      }),
    ).toBeVisible()
  })

  it('validates a task and creates it with real client-side state', async () => {
    const user = userEvent.setup()
    showLab()
    await screen.findByRole(
      'button',
      { name: /^Review workspace invitation permissions/ },
      { timeout: 2500 },
    )
    await user.click(screen.getByRole('button', { name: 'New task' }))
    const dialog = screen.getByRole('dialog', { name: 'Create task' })
    await user.click(
      within(dialog).getByRole('button', { name: 'Create task' }),
    )
    expect(
      within(dialog).getByText('Enter at least 3 characters.'),
    ).toBeVisible()
    expect(within(dialog).getByText('Enter an assignee.')).toBeVisible()
    await user.type(
      within(dialog).getByRole('textbox', { name: /^Title/ }),
      'Review release readiness',
    )
    await user.type(
      within(dialog).getByRole('textbox', { name: /^Assignee/ }),
      'Maya Iyer',
    )
    await user.click(
      within(dialog).getByRole('button', { name: 'Create task' }),
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /^Review release readiness/ }),
    ).toBeVisible()
  })

  it('undoes a deleted task within the undo window', async () => {
    const user = userEvent.setup()
    showLab()
    await screen.findByRole(
      'button',
      { name: /^Review workspace invitation permissions/ },
      { timeout: 2500 },
    )
    await user.click(
      screen.getByRole('button', {
        name: 'Delete Review workspace invitation permissions',
      }),
    )
    expect(
      screen.queryByRole('button', {
        name: /^Review workspace invitation permissions/,
      }),
    ).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Undo' }))
    expect(
      screen.getByRole('button', {
        name: /^Review workspace invitation permissions/,
      }),
    ).toBeVisible()
  })
})
