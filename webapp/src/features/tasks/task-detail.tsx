import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '../../api/client'
import { ApiError, errorMessage } from '../../api/errors'
import type { Label, Membership, Project, TaskRecord } from '../../api/types'
import { LoadingScreen } from '../../ui/states'
import { TaskForm } from './task-form'
import { taskPayload } from './task-utils'

interface TaskDetailProps {
  workspaceId: string
  taskId: string
  projects: Project[]
  members: Membership[]
  labels: Label[]
  tasks: TaskRecord[]
  canWrite: boolean
  onClose(): void
}

export function TaskDetail({
  workspaceId,
  taskId,
  projects,
  members,
  labels,
  tasks,
  canWrite,
  onClose,
}: TaskDetailProps) {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState(false)
  const [error, setError] = useState('')
  const [conflictInput, setConflictInput] = useState<Record<
    string,
    unknown
  > | null>(null)
  const task = useQuery({
    queryKey: ['task', workspaceId, taskId],
    queryFn: ({ signal }) => api.getTask(workspaceId, taskId, true, signal),
  })

  const refreshLists = async () => {
    await queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] })
  }
  const update = useMutation({
    mutationFn: (input: Record<string, unknown>) =>
      api.updateTask(workspaceId, taskId, input, crypto.randomUUID()),
    onSuccess: async (updated) => {
      queryClient.setQueryData(['task', workspaceId, taskId], updated)
      setConflictInput(null)
      setEditing(false)
      await refreshLists()
    },
    onError: (cause, input) => {
      setError(errorMessage(cause))
      if (cause instanceof ApiError && cause.status === 409)
        setConflictInput(input)
    },
  })
  const archive = useMutation({
    mutationFn: () =>
      api.archiveTask(
        workspaceId,
        taskId,
        task.data!.version,
        crypto.randomUUID(),
      ),
    onSuccess: async () => {
      await refreshLists()
      onClose()
    },
    onError: (cause) => setError(errorMessage(cause)),
  })
  const restore = useMutation({
    mutationFn: () =>
      api.restoreTask(
        workspaceId,
        taskId,
        task.data!.version,
        crypto.randomUUID(),
      ),
    onSuccess: async (restored) => {
      queryClient.setQueryData(['task', workspaceId, taskId], restored)
      await refreshLists()
    },
    onError: (cause) => setError(errorMessage(cause)),
  })

  async function applyLatest() {
    if (!conflictInput) return
    try {
      const latest = await api.getTask(workspaceId, taskId, true)
      update.mutate({ ...conflictInput, version: latest.version })
    } catch (cause) {
      setError(errorMessage(cause))
    }
  }

  if (task.isLoading)
    return (
      <div className="dialog-backdrop">
        <section className="dialog-card task-detail">
          <LoadingScreen label="Loading task" />
        </section>
      </div>
    )
  if (!task.data) return null
  const current = task.data

  return (
    <div className="dialog-backdrop" role="presentation" onMouseDown={onClose}>
      <section
        className="dialog-card task-detail"
        role="dialog"
        aria-modal="true"
        aria-labelledby="task-detail-title"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="dialog-heading">
          <div>
            <p className="eyebrow">Task detail</p>
            <h2 id="task-detail-title">{current.title}</h2>
          </div>
          <button
            className="icon-button"
            type="button"
            onClick={onClose}
            aria-label="Close task"
          >
            ×
          </button>
        </div>
        {conflictInput && (
          <div className="conflict-panel" role="alert">
            <div>
              <strong>This task changed elsewhere.</strong>
              <p>
                Reload the server version, or intentionally apply your edits to
                the latest version.
              </p>
            </div>
            <button
              className="button secondary"
              type="button"
              onClick={() => {
                setConflictInput(null)
                setEditing(false)
                void task.refetch()
              }}
            >
              Reload
            </button>
            <button
              className="button primary"
              type="button"
              onClick={() => void applyLatest()}
            >
              Apply my edits
            </button>
          </div>
        )}
        {error && !conflictInput && (
          <div className="notice error" role="alert">
            {error}
          </div>
        )}
        {editing ? (
          <TaskForm
            task={current}
            projects={projects}
            members={members}
            labels={labels}
            tasks={tasks}
            pending={update.isPending}
            error=""
            onCancel={() => setEditing(false)}
            onSubmit={(input) => update.mutate(input)}
          />
        ) : (
          <div className="task-detail-content">
            <div className="detail-summary">
              <span className={`status-chip ${current.status}`}>
                {current.status.replace('_', ' ')}
              </span>
              <span className={`priority ${current.priority}`}>
                {current.priority} priority
              </span>
              {current.dueDate && (
                <span>
                  Due{' '}
                  {new Intl.DateTimeFormat().format(new Date(current.dueDate))}
                </span>
              )}
            </div>
            <section>
              <h3>Description</h3>
              <p className="task-description">
                {current.description || 'No description provided.'}
              </p>
            </section>
            <div className="detail-grid">
              <div>
                <small>Assignee</small>
                <strong>
                  {members.find((member) => member.userId === current.assigneeId)
                    ?.userName || current.assigneeId || 'Unassigned'}
                </strong>
              </div>
              <div>
                <small>Estimate</small>
                <strong>
                  {current.estimateMinutes
                    ? `${current.estimateMinutes} minutes`
                    : 'Not estimated'}
                </strong>
              </div>
              <div>
                <small>Version</small>
                <strong>{current.version}</strong>
              </div>
              <div>
                <small>Updated</small>
                <strong>
                  {new Intl.DateTimeFormat().format(
                    new Date(current.updatedAt),
                  )}
                </strong>
              </div>
            </div>
            {current.checklist?.length ? (
              <section>
                <h3>Checklist</h3>
                <div className="checklist">
                  {current.checklist.map((item) => (
                    <label key={item.id}>
                      <input
                        type="checkbox"
                        checked={item.completed}
                        disabled={!canWrite || update.isPending}
                        onChange={() =>
                          update.mutate(
                            taskPayload(current, {
                              checklist: current.checklist!.map((entry) =>
                                entry.id === item.id
                                  ? { ...entry, completed: !entry.completed }
                                  : entry,
                              ),
                            }),
                          )
                        }
                      />
                      <span>{item.text}</span>
                    </label>
                  ))}
                </div>
              </section>
            ) : null}
            {current.labelIds?.length ? (
              <section>
                <h3>Labels</h3>
                <div className="tag-list">
                  {current.labelIds.map((id) => {
                    const label = labels.find((item) => item.id === id)
                    return (
                      <span key={id} style={{ borderColor: label?.color }}>
                        {label?.name ?? id}
                      </span>
                    )
                  })}
                </div>
              </section>
            ) : null}
            {current.dependencies?.length ? (
              <section>
                <h3>Dependencies</h3>
                <div className="tag-list">
                  {current.dependencies.map((dependency) => (
                    <span key={dependency.dependsOnTaskId}>
                      {tasks.find(
                        (item) => item.id === dependency.dependsOnTaskId,
                      )?.title ?? dependency.dependsOnTaskId}
                    </span>
                  ))}
                </div>
              </section>
            ) : null}
            <div className="dialog-actions split">
              {canWrite &&
                (current.deletedAt ? (
                  <button
                    className="button secondary"
                    type="button"
                    onClick={() => restore.mutate()}
                    disabled={restore.isPending}
                  >
                    Restore task
                  </button>
                ) : (
                  <button
                    className="button danger"
                    type="button"
                    onClick={() => archive.mutate()}
                    disabled={archive.isPending}
                  >
                    Archive
                  </button>
                ))}
              <span />
              {canWrite && !current.deletedAt && (
                <button
                  className="button primary"
                  type="button"
                  onClick={() => setEditing(true)}
                >
                  Edit task
                </button>
              )}
            </div>
          </div>
        )}
      </section>
    </div>
  )
}
