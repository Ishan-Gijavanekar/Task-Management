import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'
import { TaskForm } from './task-form'

const project = {
  id: 'project-1',
  workspaceId: 'workspace-1',
  key: 'WEB',
  name: 'Website',
  description: '',
  visibility: 'workspace' as const,
  workflow: [],
  version: 1,
}

const labels = [
  {
    id: 'label-1',
    workspaceId: 'workspace-1',
    name: 'Bug',
    color: '#bd433f',
    createdAt: '2026-11-22T00:00:00Z',
  },
  {
    id: 'label-2',
    workspaceId: 'workspace-1',
    name: 'Frontend',
    color: '#4f78df',
    createdAt: '2026-11-22T00:00:00Z',
  },
]

describe('TaskForm labels', () => {
  it('submits every checked workspace label', async () => {
    const onSubmit = vi.fn()
    render(
      <MemoryRouter>
        <TaskForm
          projectId={project.id}
          projects={[project]}
          members={[]}
          labels={labels}
          tasks={[]}
          pending={false}
          error=""
          onCancel={() => undefined}
          onSubmit={onSubmit}
        />
      </MemoryRouter>,
    )

    await userEvent.type(screen.getByLabelText('Title'), 'Fix navigation')
    await userEvent.click(screen.getByText('Select labels'))
    await userEvent.type(screen.getByLabelText('Search labels'), 'front')
    expect(screen.queryByLabelText('Bug')).not.toBeInTheDocument()
    await userEvent.click(screen.getByLabelText('Frontend'))
    await userEvent.clear(screen.getByLabelText('Search labels'))
    await userEvent.click(screen.getByLabelText('Bug'))
    expect(
      screen.getByRole('button', { name: 'Remove Frontend' }),
    ).toBeVisible()
    expect(screen.getByRole('button', { name: 'Remove Bug' })).toBeVisible()
    await userEvent.click(screen.getByRole('button', { name: 'Create task' }))

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ labelIds: ['label-2', 'label-1'] }),
    )
  })
})
