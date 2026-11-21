import { describe, expect, it } from 'vitest'
import type { TaskRecord } from '../../api/types'
import { replaceTask, taskPayload } from './task-utils'

const task: TaskRecord = {
  id: 'task-1',
  workspaceId: 'workspace-1',
  projectId: 'project-1',
  reporterId: 'user-1',
  title: 'Ship the board',
  description: 'Keep every editable field intact.',
  status: 'backlog',
  priority: 'high',
  assigneeId: 'user-2',
  labelIds: ['label-1'],
  checklist: [
    { id: 'check-1', text: 'Verify', completed: false, position: 1024 },
  ],
  dependencies: [{ taskId: 'task-1', dependsOnTaskId: 'task-blocker' }],
  startDate: '2026-11-20T00:00:00.000Z',
  dueDate: '2026-11-24T00:00:00.000Z',
  estimateMinutes: 90,
  position: 1024,
  version: 7,
  createdAt: '2026-11-20T00:00:00.000Z',
  updatedAt: '2026-11-20T00:00:00.000Z',
}

describe('task update helpers', () => {
  it('builds a complete versioned update when one field changes', () => {
    expect(taskPayload(task, { status: 'todo' })).toEqual({
      title: task.title,
      description: task.description,
      status: 'todo',
      priority: task.priority,
      assigneeId: task.assigneeId,
      labelIds: task.labelIds,
      checklist: task.checklist,
      dependencies: ['task-blocker'],
      dueDate: task.dueDate,
      startDate: task.startDate,
      estimateMinutes: task.estimateMinutes,
      version: 7,
    })
  })

  it('only replaces the matching task in an optimistic list', () => {
    const replacement = { ...task, title: 'Updated', version: 8 }
    expect(replaceTask(task, replacement)).toBe(replacement)
    expect(replaceTask({ ...task, id: 'task-2' }, replacement).title).toBe(
      task.title,
    )
  })
})
