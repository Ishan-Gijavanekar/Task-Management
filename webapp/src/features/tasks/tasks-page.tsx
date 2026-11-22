import {
  type InfiniteData,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useMemo, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api } from '../../api/client'
import { ApiError, errorMessage } from '../../api/errors'
import type { Label, TaskPage, TaskRecord } from '../../api/types'
import { EmptyState, LoadingScreen } from '../../ui/states'
import { useWorkspacePermission } from '../members/permissions'
import { useWorkspace } from '../workspaces/workspace-context'
import { TaskDetail } from './task-detail'
import { TaskForm } from './task-form'
import {
  replaceTask,
  statusLabels,
  taskPayload,
  type TaskStatus,
} from './task-utils'

const transitions: Record<TaskStatus, TaskStatus[]> = {
  backlog: ['todo', 'cancelled'],
  todo: ['backlog', 'in_progress', 'blocked', 'cancelled'],
  in_progress: ['todo', 'blocked', 'done', 'cancelled'],
  blocked: ['todo', 'in_progress', 'cancelled'],
  done: ['in_progress'],
  cancelled: ['backlog'],
}

const priorityRank = { urgent: 0, high: 1, medium: 2, low: 3, none: 4 }

export function TasksPage() {
  const [search] = useSearchParams()
  return (
    <TaskExplorer view={search.get('view') === 'board' ? 'board' : 'list'} />
  )
}

export function TaskExplorer({
  projectId,
  view,
}: {
  projectId?: string
  view: 'board' | 'list'
}) {
  const { workspace } = useWorkspace()
  const { canWrite } = useWorkspacePermission()
  const queryClient = useQueryClient()
  const [search, setSearch] = useSearchParams()
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')
  const [conflict, setConflict] = useState<{
    taskId: string
    status: TaskStatus
  } | null>(null)
  const searchTimer = useRef<number | undefined>(undefined)
  const workspaceId = workspace!.id

  const projects = useQuery({
    queryKey: ['projects', workspaceId],
    queryFn: ({ signal }) => api.listProjects(workspaceId, signal),
  })
  const members = useQuery({
    queryKey: ['members', workspaceId],
    queryFn: ({ signal }) => api.listMembers(workspaceId, signal),
  })
  const labels = useQuery({
    queryKey: ['labels', workspaceId],
    queryFn: ({ signal }) => api.listLabels(workspaceId, signal),
  })
  const taskOptions = useQuery({
    queryKey: ['task-options', workspaceId, projectId],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ limit: '200' })
      if (projectId) params.set('projectId', projectId)
      return api.listTasks(workspaceId, params, signal)
    },
  })

  const requestSearch = useMemo(() => {
    const params = new URLSearchParams()
    if (projectId) params.set('projectId', projectId)
    for (const key of ['q', 'status', 'priority', 'assigneeId', 'labelId']) {
      const value = search.get(key)
      if (value) params.set(key, value)
    }
    if (search.get('archived') === 'true') params.set('includeDeleted', 'true')
    params.set('limit', '30')
    return params
  }, [projectId, search])

  const tasks = useInfiniteQuery({
    queryKey: ['tasks', workspaceId, requestSearch.toString()],
    queryFn: ({ pageParam, signal }) => {
      const params = new URLSearchParams(requestSearch)
      if (pageParam) params.set('cursor', pageParam)
      return api.listTasks(workspaceId, params, signal)
    },
    initialPageParam: '',
    getNextPageParam: (lastPage) => lastPage.nextCursor || undefined,
  })

  const allTasks = useMemo(() => {
    const items = tasks.data?.pages.flatMap((page) => page.tasks) ?? []
    const sort = search.get('sort') ?? 'position'
    return [...items].sort((left, right) => {
      if (sort === 'due')
        return (left.dueDate ?? '9999').localeCompare(right.dueDate ?? '9999')
      if (sort === 'priority')
        return priorityRank[left.priority] - priorityRank[right.priority]
      if (sort === 'title') return left.title.localeCompare(right.title)
      return left.position - right.position
    })
  }, [search, tasks.data])

  const replaceInLists = (replacement: TaskRecord) => {
    queryClient.setQueriesData<InfiniteData<TaskPage>>(
      { queryKey: ['tasks', workspaceId] },
      (data) =>
        data
          ? {
              ...data,
              pages: data.pages.map((page) => ({
                ...page,
                tasks: page.tasks.map((task) => replaceTask(task, replacement)),
              })),
            }
          : data,
    )
  }

  const move = useMutation({
    mutationFn: ({ task, status }: { task: TaskRecord; status: TaskStatus }) =>
      api.updateTask(
        workspaceId,
        task.id,
        taskPayload(task, { status }),
        crypto.randomUUID(),
      ),
    onMutate: async ({ task, status }) => {
      await queryClient.cancelQueries({ queryKey: ['tasks', workspaceId] })
      const snapshots = queryClient.getQueriesData<InfiniteData<TaskPage>>({
        queryKey: ['tasks', workspaceId],
      })
      replaceInLists({ ...task, status })
      return snapshots
    },
    onSuccess: (updated) => replaceInLists(updated),
    onError: (cause, variables, snapshots) => {
      snapshots?.forEach(([key, data]) => queryClient.setQueryData(key, data))
      setError(errorMessage(cause))
      if (cause instanceof ApiError && cause.status === 409)
        setConflict({ taskId: variables.task.id, status: variables.status })
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] }),
  })

  const create = useMutation({
    mutationFn: (input: Record<string, unknown>) =>
      api.createTask(workspaceId, input, crypto.randomUUID()),
    onSuccess: async (created) => {
      setCreating(false)
      await queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] })
      setParam('task', created.id)
    },
    onError: (cause) => setError(errorMessage(cause)),
  })

  function setParam(key: string, value?: string) {
    setSearch(
      (current) => {
        const next = new URLSearchParams(current)
        if (value) next.set(key, value)
        else next.delete(key)
        return next
      },
      { replace: true },
    )
  }

  function moveTask(task: TaskRecord, status: TaskStatus) {
    if (task.status === status) return
    if (!transitions[task.status].includes(status)) {
      setError(
        `Tasks cannot move directly from ${statusLabels[task.status]} to ${statusLabels[status]}.`,
      )
      return
    }
    setError('')
    move.mutate({ task, status })
  }

  async function applyConflict() {
    if (!conflict) return
    try {
      const latest = await api.getTask(workspaceId, conflict.taskId)
      const updated = await api.updateTask(
        workspaceId,
        latest.id,
        taskPayload(latest, { status: conflict.status }),
        crypto.randomUUID(),
      )
      replaceInLists(updated)
      setConflict(null)
      setError('')
    } catch (cause) {
      setError(errorMessage(cause))
    }
  }

  function saveFilter() {
    const name = window.prompt('Name this filter')?.trim()
    if (!name) return
    const key = `taskflow.filters.${workspaceId}`
    const saved = JSON.parse(localStorage.getItem(key) ?? '{}') as Record<
      string,
      string
    >
    saved[name] = search.toString()
    localStorage.setItem(key, JSON.stringify(saved))
  }

  const currentProject = projects.data?.projects.find(
    (project) => project.id === projectId,
  )
  const detailId = search.get('task')
  const statuses = Object.keys(statusLabels) as TaskStatus[]
  const savedFilters = JSON.parse(
    localStorage.getItem(`taskflow.filters.${workspaceId}`) ?? '{}',
  ) as Record<string, string>

  return (
    <div className="page task-page">
      <header className="page-header task-header">
        <div>
          <p className="eyebrow">{currentProject?.key ?? 'All work'}</p>
          <h1>{currentProject?.name ?? 'Tasks'}</h1>
          <p>
            {currentProject?.description ||
              'Search, filter, and move work forward.'}
          </p>
        </div>
        <div className="header-actions">
          {projectId && canWrite && (
            <Link className="button secondary" to="/projects?new=true">
              New project
            </Link>
          )}
          {canWrite && projects.data?.projects.length ? (
            <button
              className="button primary"
              type="button"
              onClick={() => setCreating(true)}
            >
              ＋ New task
            </button>
          ) : null}
        </div>
      </header>

      {projectId && (
        <nav className="view-tabs" aria-label="Project views">
          <Link
            className={view === 'board' ? 'active' : ''}
            to={`/projects/${projectId}/board?${search.toString()}`}
          >
            Board
          </Link>
          <Link
            className={view === 'list' ? 'active' : ''}
            to={`/projects/${projectId}/list?${search.toString()}`}
          >
            List
          </Link>
          <Link to={`/projects/${projectId}/settings`}>Settings</Link>
        </nav>
      )}

      <section className="task-toolbar" aria-label="Task filters">
        <label className="search-field">
          <span aria-hidden="true">⌕</span>
          <span className="sr-only">Search tasks</span>
          <input
            key={search.get('q')}
            defaultValue={search.get('q') ?? ''}
            placeholder="Search tasks…"
            onChange={(event) => {
              window.clearTimeout(searchTimer.current)
              const value = event.currentTarget.value
              searchTimer.current = window.setTimeout(
                () => setParam('q', value),
                350,
              )
            }}
          />
        </label>
        <select
          aria-label="Filter by status"
          value={search.get('status') ?? ''}
          onChange={(event) => setParam('status', event.target.value)}
        >
          <option value="">All statuses</option>
          {statuses.map((status) => (
            <option key={status} value={status}>
              {statusLabels[status]}
            </option>
          ))}
        </select>
        <select
          aria-label="Filter by priority"
          value={search.get('priority') ?? ''}
          onChange={(event) => setParam('priority', event.target.value)}
        >
          <option value="">All priorities</option>
          <option value="urgent">Urgent</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
          <option value="none">None</option>
        </select>
        <select
          aria-label="Filter by assignee"
          value={search.get('assigneeId') ?? ''}
          onChange={(event) => setParam('assigneeId', event.target.value)}
        >
          <option value="">All assignees</option>
          {members.data?.members
            .filter((member) => member.status === 'active')
            .map((member) => (
              <option key={member.userId} value={member.userId}>
                {member.userName || member.userId}
              </option>
            ))}
        </select>
        <select
          aria-label="Sort tasks"
          value={search.get('sort') ?? 'position'}
          onChange={(event) => setParam('sort', event.target.value)}
        >
          <option value="position">Board order</option>
          <option value="due">Due date</option>
          <option value="priority">Priority</option>
          <option value="title">Title</option>
        </select>
        <select
          aria-label="Filter by label"
          value={search.get('labelId') ?? ''}
          onChange={(event) => setParam('labelId', event.target.value)}
        >
          <option value="">All labels</option>
          {labels.data?.labels.map((label) => (
            <option key={label.id} value={label.id}>
              {label.name}
            </option>
          ))}
        </select>
        {canWrite && (
          <Link className="button secondary compact" to="/settings/labels">
            Manage labels
          </Link>
        )}
        {Object.keys(savedFilters).length ? (
          <select
            aria-label="Open saved filter"
            defaultValue=""
            onChange={(event) => {
              const saved = savedFilters[event.target.value]
              if (saved !== undefined) setSearch(saved, { replace: true })
              event.target.value = ''
            }}
          >
            <option value="">Saved filters</option>
            {Object.keys(savedFilters).map((name) => (
              <option key={name} value={name}>
                {name}
              </option>
            ))}
          </select>
        ) : null}
        <button
          className="button secondary compact"
          type="button"
          onClick={saveFilter}
        >
          Save filter
        </button>
        <button
          className={`button secondary compact ${search.get('archived') === 'true' ? 'active' : ''}`}
          type="button"
          onClick={() =>
            setParam(
              'archived',
              search.get('archived') === 'true' ? undefined : 'true',
            )
          }
        >
          Archived
        </button>
      </section>

      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      {conflict && (
        <div className="conflict-panel">
          <div>
            <strong>Someone updated this task first.</strong>
            <p>
              Reload their change or intentionally apply your move to the newest
              version.
            </p>
          </div>
          <button
            className="button secondary"
            type="button"
            onClick={() => {
              setConflict(null)
              setError('')
              void tasks.refetch()
            }}
          >
            Reload
          </button>
          <button
            className="button primary"
            type="button"
            onClick={() => void applyConflict()}
          >
            Apply move
          </button>
        </div>
      )}

      {tasks.isLoading ? (
        <LoadingScreen label="Loading tasks" />
      ) : !allTasks.length ? (
        <div className="panel">
          <EmptyState title="No tasks match">
            Adjust the filters or create a task to get work moving.
          </EmptyState>
        </div>
      ) : view === 'board' ? (
        <div className="task-board">
          {statuses
            .filter(
              (status) =>
                status !== 'cancelled' ||
                allTasks.some((task) => task.status === status),
            )
            .map((status) => (
              <section
                className="board-column"
                key={status}
                onDragOver={(event) => event.preventDefault()}
                onDrop={(event) => {
                  const task = allTasks.find(
                    (item) =>
                      item.id === event.dataTransfer.getData('text/task-id'),
                  )
                  if (task) moveTask(task, status)
                }}
              >
                <header>
                  <span className={`status-dot ${status}`} />
                  <h2>{statusLabels[status]}</h2>
                  <span>
                    {allTasks.filter((task) => task.status === status).length}
                  </span>
                </header>
                <div className="board-stack">
                  {allTasks
                    .filter((task) => task.status === status)
                    .map((task) => (
                      <TaskCard
                        key={task.id}
                        task={task}
                        labels={labels.data?.labels ?? []}
                        canWrite={canWrite}
                        onOpen={() => setParam('task', task.id)}
                        onMove={(next) => moveTask(task, next)}
                      />
                    ))}
                </div>
              </section>
            ))}
        </div>
      ) : (
        <div className="task-list panel">
          <div className="task-list-head">
            <span>Task</span>
            <span>Status</span>
            <span>Priority</span>
            <span>Due</span>
            <span>Assignee</span>
          </div>
          {allTasks.map((task) => (
            <button
              className="task-list-row"
              key={task.id}
              type="button"
              onClick={() => setParam('task', task.id)}
            >
              <span>
                <strong>{task.title}</strong>
                <small>
                  {
                    projects.data?.projects.find(
                      (project) => project.id === task.projectId,
                    )?.key
                  }
                </small>
                <TaskLabels
                  labelIds={task.labelIds}
                  labels={labels.data?.labels ?? []}
                />
              </span>
              <span className={`status-chip ${task.status}`}>
                {statusLabels[task.status]}
              </span>
              <span className={`priority ${task.priority}`}>
                {task.priority}
              </span>
              <span>
                {task.dueDate
                  ? new Intl.DateTimeFormat().format(new Date(task.dueDate))
                  : '—'}
              </span>
              <span>
                {members.data?.members.find(
                  (member) => member.userId === task.assigneeId,
                )?.userName ||
                  task.assigneeId ||
                  'Unassigned'}
              </span>
            </button>
          ))}
        </div>
      )}

      {tasks.hasNextPage && (
        <div className="load-more">
          <button
            className="button secondary"
            type="button"
            onClick={() => void tasks.fetchNextPage()}
            disabled={tasks.isFetchingNextPage}
          >
            {tasks.isFetchingNextPage ? 'Loading…' : 'Load more'}
          </button>
        </div>
      )}

      {creating && (
        <div
          className="dialog-backdrop"
          role="presentation"
          onMouseDown={() => setCreating(false)}
        >
          <section
            className="dialog-card task-detail"
            role="dialog"
            aria-modal="true"
            aria-labelledby="new-task-title"
            onMouseDown={(event) => event.stopPropagation()}
          >
            <div className="dialog-heading">
              <div>
                <p className="eyebrow">New task</p>
                <h2 id="new-task-title">Create a clear next step</h2>
              </div>
              <button
                className="icon-button"
                type="button"
                onClick={() => setCreating(false)}
                aria-label="Close"
              >
                ×
              </button>
            </div>
            <TaskForm
              projectId={projectId}
              projects={projects.data?.projects ?? []}
              members={members.data?.members ?? []}
              labels={labels.data?.labels ?? []}
              tasks={taskOptions.data?.tasks ?? []}
              pending={create.isPending}
              error={error}
              onCancel={() => setCreating(false)}
              onSubmit={(input) => create.mutate(input)}
            />
          </section>
        </div>
      )}
      {detailId && (
        <TaskDetail
          workspaceId={workspaceId}
          taskId={detailId}
          projects={projects.data?.projects ?? []}
          members={members.data?.members ?? []}
          labels={labels.data?.labels ?? []}
          tasks={taskOptions.data?.tasks ?? []}
          canWrite={canWrite}
          onClose={() => setParam('task')}
        />
      )}
    </div>
  )
}

function TaskCard({
  task,
  labels,
  canWrite,
  onOpen,
  onMove,
}: {
  task: TaskRecord
  labels: Label[]
  canWrite: boolean
  onOpen(): void
  onMove(status: TaskStatus): void
}) {
  return (
    <article
      className="task-card"
      draggable={canWrite}
      onDragStart={(event) => {
        event.dataTransfer.effectAllowed = 'move'
        event.dataTransfer.setData('text/task-id', task.id)
      }}
    >
      <button className="task-card-main" type="button" onClick={onOpen}>
        <span className={`priority-bar ${task.priority}`} />
        <strong>{task.title}</strong>
        <TaskLabels labelIds={task.labelIds} labels={labels} />
        {task.description && <p>{task.description}</p>}
        <div className="task-meta">
          {task.dueDate && (
            <span>
              ◷{' '}
              {new Intl.DateTimeFormat(undefined, {
                month: 'short',
                day: 'numeric',
              }).format(new Date(task.dueDate))}
            </span>
          )}
          {task.checklist?.length ? (
            <span>
              ☑ {task.checklist.filter((item) => item.completed).length}/
              {task.checklist.length}
            </span>
          ) : null}
          {task.dependencies?.length ? (
            <span>↗ {task.dependencies.length}</span>
          ) : null}
        </div>
      </button>
      {canWrite && (
        <label className="move-select">
          <span className="sr-only">Move {task.title}</span>
          <select
            aria-label={`Move ${task.title}`}
            value={task.status}
            onChange={(event) => onMove(event.target.value as TaskStatus)}
          >
            <option value={task.status}>{statusLabels[task.status]}</option>
            {transitions[task.status].map((status) => (
              <option key={status} value={status}>
                {statusLabels[status]}
              </option>
            ))}
          </select>
        </label>
      )}
    </article>
  )
}

function TaskLabels({
  labelIds = [],
  labels,
}: {
  labelIds: string[] | undefined
  labels: Label[]
}) {
  const assigned = labelIds
    .map((id) => labels.find((label) => label.id === id))
    .filter((label) => label !== undefined)
  if (!assigned.length) return null
  return (
    <span className="task-labels">
      {assigned.map((label) => (
        <span key={label.id}>
          <i style={{ backgroundColor: label.color }} />
          {label.name}
        </span>
      ))}
    </span>
  )
}
