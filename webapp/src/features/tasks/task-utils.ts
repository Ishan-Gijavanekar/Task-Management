import type { TaskRecord } from '../../api/types'

export const statusLabels = {
  backlog: 'Backlog',
  todo: 'To do',
  in_progress: 'In progress',
  blocked: 'Blocked',
  done: 'Done',
  cancelled: 'Cancelled',
} as const

export type TaskStatus = keyof typeof statusLabels

export function taskPayload(
  task: TaskRecord,
  changes: Partial<TaskRecord> = {},
) {
  const next = { ...task, ...changes }
  return {
    title: next.title,
    description: next.description ?? '',
    status: next.status,
    priority: next.priority,
    assigneeId: next.assigneeId ?? '',
    labelIds: next.labelIds ?? [],
    checklist: next.checklist ?? [],
    dependencies: next.dependencies?.map((item) => item.dependsOnTaskId) ?? [],
    dueDate: next.dueDate,
    startDate: next.startDate,
    estimateMinutes: next.estimateMinutes ?? 0,
    version: task.version,
  }
}

export function replaceTask(task: TaskRecord, replacement: TaskRecord) {
  return task.id === replacement.id ? replacement : task
}
