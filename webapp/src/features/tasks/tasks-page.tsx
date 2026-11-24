import './tasks.css'
import './board-refresh.css'
import {
  type InfiniteData,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useEffect, useMemo, useRef, useState } from 'react'
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
  const searchInputRef = useRef<HTMLInputElement>(null)
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [busyBulk, setBusyBulk] = useState(false)
  const [undoTasks, setUndoTasks] = useState<TaskRecord[] | null>(null)
  const [toastMessage, setToastMessage] = useState('')
  const toastTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
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
    const ordered = [...items].sort((left, right) => {
      if (sort === 'due')
        return (left.dueDate ?? '9999').localeCompare(right.dueDate ?? '9999')
      if (sort === 'priority')
        return priorityRank[left.priority] - priorityRank[right.priority]
      if (sort === 'title') return left.title.localeCompare(right.title)
      if (sort === 'status') return left.status.localeCompare(right.status)
      return left.position - right.position
    })
    return search.get('direction') === 'desc' ? ordered.reverse() : ordered
  }, [search, tasks.data])

  const replaceInLists = (replacement: TaskRecord) => {
    queryClient.setQueriesData<InfiniteData<TaskPage>>(
      { queryKey: ['tasks', workspaceId] },
      (data) =>
        data && Array.isArray(data.pages)
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

  const currentProject = projects.data?.projects.find(
    (project) => project.id === projectId,
  )
  const detailId = search.get('task')
  const statuses = Object.keys(statusLabels) as TaskStatus[]
  const selectedTasks = allTasks.filter((task) => selectedIds.includes(task.id))
  const allSelected =
    allTasks.length > 0 && selectedTasks.length === allTasks.length
  const activeFilters = [
    'q',
    'status',
    'priority',
    'assigneeId',
    'labelId',
    'archived',
  ].filter((key) => search.get(key))

  useEffect(() => {
    function shortcut(event: KeyboardEvent) {
      const element = event.target as HTMLElement | null
      const typing =
        element?.matches('input, textarea, select, [contenteditable="true"]') ??
        false
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        searchInputRef.current?.focus()
      }
      if (event.key === 'Escape') {
        setCreating(false)
        setSearch(
          (current) => {
            if (!current.has('task')) return current
            const next = new URLSearchParams(current)
            next.delete('task')
            return next
          },
          { replace: true },
        )
      }
      if (
        event.key.toLowerCase() === 'n' &&
        !typing &&
        !event.ctrlKey &&
        !event.metaKey &&
        !event.altKey &&
        canWrite &&
        projects.data?.projects.length
      ) {
        event.preventDefault()
        setCreating(true)
      }
    }
    window.addEventListener('keydown', shortcut)
    return () => window.removeEventListener('keydown', shortcut)
  }, [canWrite, projects.data?.projects.length, setSearch])

  useEffect(
    () => () => {
      window.clearTimeout(searchTimer.current)
      if (toastTimeout.current) clearTimeout(toastTimeout.current)
    },
    [],
  )

  function sortBy(column: 'title' | 'status' | 'priority' | 'due') {
    setSearch(
      (current) => {
        const next = new URLSearchParams(current)
        const same = next.get('sort') === column
        next.set('sort', column)
        next.set(
          'direction',
          same && next.get('direction') !== 'desc' ? 'desc' : 'asc',
        )
        return next
      },
      { replace: true },
    )
  }

  async function changeSelected(status: TaskStatus) {
    const eligible = selectedTasks.filter((task) =>
      transitions[task.status].includes(status),
    )
    const skipped = selectedTasks.length - eligible.length
    setBusyBulk(true)
    try {
      const outcomes = await Promise.allSettled(
        eligible.map((task) =>
          api.updateTask(
            workspaceId,
            task.id,
            taskPayload(task, { status }),
            crypto.randomUUID(),
          ),
        ),
      )
      const failed = outcomes.filter(
        (result) => result.status === 'rejected',
      ).length
      setError(
        failed || skipped
          ? `${failed} updates failed; ${skipped} transitions not permitted. Refresh to review the latest task states.`
          : '',
      )
      setSelectedIds([])
      await queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] })
    } finally {
      setBusyBulk(false)
    }
  }

  async function archiveSelected() {
    if (!selectedTasks.length || busyBulk) return
    setBusyBulk(true)
    const snapshot = queryClient.getQueriesData<InfiniteData<TaskPage>>({
      queryKey: ['tasks', workspaceId],
    })
    const ids = new Set(selectedTasks.map((task) => task.id))
    queryClient.setQueriesData<InfiniteData<TaskPage>>(
      { queryKey: ['tasks', workspaceId] },
      (data) =>
        data && Array.isArray(data.pages)
          ? {
              ...data,
              pages: data.pages.map((page) => ({
                ...page,
                tasks: page.tasks.filter((task) => !ids.has(task.id)),
              })),
            }
          : data,
    )
    setSelectedIds([])
    try {
      const outcomes = await Promise.allSettled(
        selectedTasks.map((task) =>
          api.archiveTask(
            workspaceId,
            task.id,
            task.version,
            crypto.randomUUID(),
          ),
        ),
      )
      const archived = selectedTasks.filter(
        (_, index) => outcomes[index].status === 'fulfilled',
      )
      const failed = outcomes.length - archived.length
      if (failed)
        setError(`${failed} tasks could not be archived. Task list refreshed.`)
      await queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] })
      if (archived.length) {
        setUndoTasks(archived)
        setToastMessage(
          `${archived.length} task${archived.length === 1 ? '' : 's'} archived`,
        )
        if (toastTimeout.current) clearTimeout(toastTimeout.current)
        toastTimeout.current = setTimeout(() => {
          setUndoTasks(null)
          setToastMessage('')
        }, 5000)
      }
    } catch (cause) {
      snapshot.forEach(([key, data]) => queryClient.setQueryData(key, data))
      setError(errorMessage(cause))
    } finally {
      setBusyBulk(false)
    }
  }

  async function undoArchive() {
    if (!undoTasks) return
    const target = undoTasks
    setUndoTasks(null)
    if (toastTimeout.current) clearTimeout(toastTimeout.current)
    setToastMessage('Restoring tasks…')
    const outcomes = await Promise.allSettled(
      target.map(async (task) => {
        const latest = await api.getTask(workspaceId, task.id, true)
        return api.restoreTask(
          workspaceId,
          task.id,
          latest.version,
          crypto.randomUUID(),
        )
      }),
    )
    const failed = outcomes.filter(
      (result) => result.status === 'rejected',
    ).length
    setToastMessage(
      failed
        ? `${failed} restores failed. Check the task list.`
        : 'Tasks restored',
    )
    await queryClient.invalidateQueries({ queryKey: ['tasks', workspaceId] })
  }

  return (
    <div
      className={`page task-page ${view === 'board' ? 'task-page--board' : ''}`}
    >
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
              Project
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

      <section
        className="task-toolbar task-toolbar--compact"
        aria-label="Task search and filters"
      >
        <div className="task-search">
          <svg
            viewBox="0 0 24 24"
            width="19"
            height="19"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <circle cx="11" cy="11" r="7" />
            <path d="m16 16 4 4" />
          </svg>

          <input
            ref={searchInputRef}
            type="search"
            aria-label="Search tasks"
            key={search.get('q') ?? ''}
            defaultValue={search.get('q') ?? ''}
            placeholder="Search tasks by title or description..."
            onChange={(event) => {
              window.clearTimeout(searchTimer.current)
              const value = event.target.value

              searchTimer.current = window.setTimeout(() => {
                setParam('q', value.trim() || undefined)
              }, 350)
            }}
          />
        </div>

        <details className="task-filter-dropdown">
          <summary>
            <span>Filters</span>
            {['status', 'priority', 'assigneeId', 'labelId', 'archived'].filter(
              (key) => search.has(key),
            ).length > 0 && (
              <span className="filter-count">
                {
                  [
                    'status',
                    'priority',
                    'assigneeId',
                    'labelId',
                    'archived',
                  ].filter((key) => search.has(key)).length
                }
              </span>
            )}
          </summary>

          <div className="task-filter-panel">
            <label>
              Status
              <select
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
            </label>

            <label>
              Priority
              <select
                value={search.get('priority') ?? ''}
                onChange={(event) => setParam('priority', event.target.value)}
              >
                <option value="">All priorities</option>
                {['urgent', 'high', 'medium', 'low', 'none'].map((priority) => (
                  <option key={priority} value={priority}>
                    {priority}
                  </option>
                ))}
              </select>
            </label>

            <label>
              Assignee
              <select
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
            </label>

            <label>
              Label
              <select
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
            </label>

            <label>
              Sort by
              <select
                value={search.get('sort') ?? 'position'}
                onChange={(event) => setParam('sort', event.target.value)}
              >
                <option value="position">Board order</option>
                <option value="due">Due date</option>
                <option value="priority">Priority</option>
                <option value="title">Title</option>
                <option value="status">Status</option>
              </select>
            </label>

            <label className="filter-checkbox">
              <input
                type="checkbox"
                checked={search.get('archived') === 'true'}
                onChange={(event) =>
                  setParam(
                    'archived',
                    event.target.checked ? 'true' : undefined,
                  )
                }
              />
              Include archived tasks
            </label>

            <button
              className="filter-clear"
              type="button"
              onClick={() => {
                const next = new URLSearchParams(search)
                for (const key of [
                  'status',
                  'priority',
                  'assigneeId',
                  'labelId',
                  'sort',
                  'direction',
                  'archived',
                ]) {
                  next.delete(key)
                }
                setSearch(next, { replace: true })
              }}
            >
              Clear filters
            </button>
          </div>
        </details>
      </section>

      {activeFilters.length > 0 && (
        <div className="task-active-filters" aria-label="Active filters">
          {activeFilters.map((key) => (
            <button
              className="task-filter-pill"
              type="button"
              key={key}
              onClick={() => setParam(key)}
            >
              {key === 'q'
                ? 'Search'
                : key === 'assigneeId'
                  ? 'Assignee'
                  : key === 'labelId'
                    ? 'Label'
                    : key === 'archived'
                      ? 'Archived'
                      : key}
              : {search.get(key)} ×
            </button>
          ))}
          <button
            className="task-filter-reset"
            type="button"
            onClick={() =>
              setSearch(
                (current) => {
                  const next = new URLSearchParams(current)
                  activeFilters.forEach((key) => next.delete(key))
                  next.delete('sort')
                  next.delete('direction')
                  return next
                },
                { replace: true },
              )
            }
          >
            Clear all
          </button>
        </div>
      )}
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
      ) : tasks.isError ? (
        <div className="notice error" role="alert">
          Tasks couldn't be loaded.{' '}
          <button
            type="button"
            className="button secondary"
            onClick={() => void tasks.refetch()}
          >
            Try again
          </button>
        </div>
      ) : !allTasks.length ? (
        <div className="panel">
          <EmptyState
            title="No tasks match"
            action={
              activeFilters.length ? (
                <button
                  className="button primary"
                  type="button"
                  onClick={() =>
                    setSearch(
                      (current) => {
                        const next = new URLSearchParams(current)
                        activeFilters.forEach((key) => next.delete(key))
                        next.delete('sort')
                        next.delete('direction')
                        return next
                      },
                      { replace: true },
                    )
                  }
                >
                  Reset filters
                </button>
              ) : canWrite && projects.data?.projects.length ? (
                <button
                  className="button primary"
                  type="button"
                  onClick={() => setCreating(true)}
                >
                  Create first task
                </button>
              ) : (
                <Link className="button secondary" to="/projects">
                  View projects
                </Link>
              )
            }
          >
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
                data-status={status}
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
                  {!allTasks.some((task) => task.status === status) && (
                    <div className="board-empty" role="status">
                      <svg
                        width="30"
                        height="30"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.3"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        aria-hidden="true"
                      >
                        <rect x="3" y="5" width="18" height="15" rx="2" />
                        <path d="M7 10h10M7 14h7" />
                      </svg>
                      <p>No tasks here</p>
                      <small>Move a task into this stage</small>
                    </div>
                  )}

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
          <div className="task-list-head-wrap">
            <input
              type="checkbox"
              aria-label="Select all loaded tasks"
              checked={allSelected}
              onChange={(event) =>
                setSelectedIds(
                  event.target.checked ? allTasks.map((task) => task.id) : [],
                )
              }
            />
            <div className="task-list-head">
              <button type="button" onClick={() => sortBy('title')}>
                Task
              </button>
              <button type="button" onClick={() => sortBy('status')}>
                Status
              </button>
              <button type="button" onClick={() => sortBy('priority')}>
                Priority
              </button>
              <button type="button" onClick={() => sortBy('due')}>
                Due
              </button>
              <span>Assignee</span>
            </div>
          </div>
          {allTasks.map((task) => (
            <div className="task-list-entry" key={task.id}>
              <input
                type="checkbox"
                aria-label={`Select ${task.title}`}
                checked={selectedIds.includes(task.id)}
                onChange={(event) =>
                  setSelectedIds((old) =>
                    event.target.checked
                      ? [...old, task.id]
                      : old.filter((id) => id !== task.id),
                  )
                }
              />
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
            </div>
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

      {canWrite && selectedTasks.length > 0 && (
        <div
          className="task-bulk-bar"
          role="toolbar"
          aria-label="Selected task actions"
        >
          <strong>{selectedTasks.length} selected</strong>
          <select
            aria-label="Change selected task status"
            defaultValue=""
            disabled={busyBulk}
            onChange={(event) => {
              if (event.target.value)
                void changeSelected(event.target.value as TaskStatus)
              event.target.value = ''
            }}
          >
            <option value="">Change status…</option>
            {statuses.map((status) => (
              <option key={status} value={status}>
                {statusLabels[status]}
              </option>
            ))}
          </select>
          <button
            type="button"
            className="button danger"
            disabled={busyBulk}
            onClick={() => void archiveSelected()}
          >
            Archive
          </button>
          <button
            type="button"
            className="button secondary"
            disabled={busyBulk}
            onClick={() => setSelectedIds([])}
          >
            Clear
          </button>
        </div>
      )}
      {toastMessage && (
        <div className="task-toast" role="status">
          <span>{toastMessage}</span>
          {undoTasks && (
            <button type="button" onClick={() => void undoArchive()}>
              Undo
            </button>
          )}
          <button
            type="button"
            aria-label="Dismiss notification"
            onClick={() => {
              setToastMessage('')
              setUndoTasks(null)
            }}
          >
            ×
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
