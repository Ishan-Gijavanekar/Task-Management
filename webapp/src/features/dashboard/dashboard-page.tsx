import { useQuery } from '@tanstack/react-query'
import { useEffect, useState, type CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../api/client'
import type { TaskRecord } from '../../api/types'
import { useAuth } from '../auth/auth-context'
import { useWorkspace } from '../workspaces/workspace-context'
import './dashboard.css'

const DAY = 86_400_000
const STATUS = [
  { key: 'backlog', label: 'Backlog', color: 'var(--dash-stone)' },
  { key: 'todo', label: 'To do', color: 'var(--dash-ochre)' },
  { key: 'in_progress', label: 'In progress', color: 'var(--dash-sage)' },
  { key: 'blocked', label: 'Blocked', color: 'var(--dash-clay)' },
  { key: 'done', label: 'Done', color: 'var(--dash-forest)' },
  { key: 'cancelled', label: 'Cancelled', color: 'var(--dash-muted)' },
] as const

function greetingFor(date: Date) {
  const hour = date.getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 17) return 'Good afternoon'
  return 'Good evening'
}

function priorityFor(task: TaskRecord, now: number) {
  if (task.dueDate && new Date(task.dueDate).getTime() < now) return 0
  if (task.status === 'blocked') return 1
  if (task.status === 'in_progress') return 2
  return 3
}

function prettyStatus(status: string) {
  return (
    STATUS.find((item) => item.key === status)?.label ??
    status.replaceAll('_', ' ')
  )
}

function relativeDue(dueDate: string | undefined, now: number) {
  if (!dueDate) return 'No due date'
  const date = new Date(dueDate)
  if (Number.isNaN(date.getTime())) return 'No due date'
  const days = Math.ceil((date.getTime() - now) / DAY)
  if (days < 0) return `${Math.abs(days)}d overdue`
  if (days === 0) return 'Due today'
  if (days === 1) return 'Due tomorrow'
  return new Intl.DateTimeFormat(undefined, {
    day: 'numeric',
    month: 'short',
  }).format(date)
}

export function DashboardPage() {
  const { user } = useAuth()
  const { workspace } = useWorkspace()
  const workspaceId = workspace!.id
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const clock = window.setInterval(() => setNow(Date.now()), 60_000)
    return () => window.clearInterval(clock)
  }, [])

  const projects = useQuery({
    queryKey: ['projects', workspaceId],
    queryFn: ({ signal }) => api.listProjects(workspaceId, signal),
  })
  const tasks = useQuery({
    queryKey: ['tasks', workspaceId, 'overview'],
    queryFn: ({ signal }) =>
      api.listTasks(workspaceId, new URLSearchParams({ limit: '50' }), signal),
  })

  const date = new Date(now)
  const today = new Intl.DateTimeFormat(undefined, {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  }).format(date)
  const firstName = user?.name?.trim().split(/\s+/)[0] || 'there'
  const items = tasks.data?.tasks ?? []
  const total = items.length
  const count = (status: string) =>
    items.filter((task) => task.status === status).length
  const done = count('done')
  const inProgress = count('in_progress')
  const blocked = count('blocked')
  const unfinished = items.filter(
    (task) => !['done', 'cancelled'].includes(task.status),
  )
  const overdue = unfinished.filter(
    (task) => task.dueDate && new Date(task.dueDate).getTime() < now,
  ).length
  const dueSoon = unfinished.filter((task) => {
    if (!task.dueDate) return false
    const due = new Date(task.dueDate).getTime()
    return due >= now && due <= now + 3 * DAY
  }).length
  const completion = total ? Math.round((done / total) * 100) : 0
  const projectNames = new Map(
    projects.data?.projects.map((project) => [project.id, project.name]) ?? [],
  )
  const focus = [...unfinished]
    .sort((a, b) => {
      const urgency = priorityFor(a, now) - priorityFor(b, now)
      if (urgency) return urgency
      return (a.dueDate ?? '9999').localeCompare(b.dueDate ?? '9999')
    })
    .slice(0, 5)

  return (
    <div className="dashboard-page page">
      <header className="dashboard-intro">
        <div>
          <p className="dashboard-overline">
            WORKSPACE OVERVIEW <span aria-hidden="true">/</span> {today}
          </p>
          <h1>
            {greetingFor(date)}, {firstName}
            <span className="dashboard-period">.</span>
          </h1>
          <p className="dashboard-subtitle">
            A clearer view of what matters in <strong>{workspace!.name}</strong>
            .
          </p>
        </div>
        <Link className="dashboard-cta" to="/projects">
          Explore projects <span aria-hidden="true">↗</span>
        </Link>
      </header>

      {tasks.isPending ? (
        <div className="dashboard-message" role="status">
          Loading your workspace overview…
        </div>
      ) : tasks.isError ? (
        <div className="dashboard-message dashboard-error" role="alert">
          Could not load tasks.{' '}
          <button type="button" onClick={() => void tasks.refetch()}>
            Try again
          </button>
        </div>
      ) : (
        <>
          <section className="dashboard-stats" aria-label="Task summary">
            <article className="dashboard-stat dashboard-stat-primary">
              <div className="dashboard-stat-head">
                <span>Tasks completed</span>
                <span className="dashboard-stat-symbol" aria-hidden="true">
                  ✓
                </span>
              </div>
              <strong>
                {done}
                <small> / {total}</small>
              </strong>
              <span className="dashboard-stat-foot">
                {completion}% of loaded tasks completed
              </span>
            </article>
            <article className="dashboard-stat">
              <div className="dashboard-stat-head">
                <span>In progress</span>
                <span className="dashboard-stat-symbol" aria-hidden="true">
                  ↗
                </span>
              </div>
              <strong>{inProgress}</strong>
              <span className="dashboard-stat-foot">
                Currently moving forward
              </span>
            </article>
            <article className="dashboard-stat">
              <div className="dashboard-stat-head">
                <span>Due in 3 days</span>
                <span className="dashboard-stat-symbol" aria-hidden="true">
                  ◷
                </span>
              </div>
              <strong>{dueSoon}</strong>
              <span className="dashboard-stat-foot">
                Upcoming, excluding overdue
              </span>
            </article>
            <article className="dashboard-stat">
              <div className="dashboard-stat-head">
                <span>Needs attention</span>
                <span className="dashboard-stat-symbol" aria-hidden="true">
                  !
                </span>
              </div>
              <strong>
                {overdue +
                  blocked -
                  unfinished.filter(
                    (task) =>
                      task.status === 'blocked' &&
                      task.dueDate &&
                      new Date(task.dueDate).getTime() < now,
                  ).length}
              </strong>
              <span className="dashboard-stat-foot">Overdue or blocked</span>
            </article>
          </section>

          <div className="dashboard-panels">
            <section
              className="dashboard-panel dashboard-focus"
              aria-labelledby="dashboard-focus-title"
            >
              <div className="dashboard-section-heading">
                <div>
                  <p className="dashboard-kicker">UP NEXT</p>
                  <h2 id="dashboard-focus-title">Work to focus on</h2>
                  <p>Open tasks, prioritized by urgency.</p>
                </div>
                <Link to="/tasks">
                  All tasks <span aria-hidden="true">↗</span>
                </Link>
              </div>
              {focus.length ? (
                <div className="dashboard-focus-list">
                  {focus.map((task) => (
                    <Link
                      key={task.id}
                      to={`/tasks?task=${encodeURIComponent(task.id)}`}
                      className="dashboard-focus-row"
                      aria-label={`Open ${task.title}`}
                    >
                      <span
                        className={`dashboard-status-marker dashboard-marker-${task.status}`}
                        aria-hidden="true"
                      />
                      <span className="dashboard-focus-title">
                        <strong>{task.title}</strong>
                        <small>
                          {projectNames.get(task.projectId) ?? 'Project'} ·{' '}
                          {prettyStatus(task.status)}
                        </small>
                      </span>
                      <span
                        className={`dashboard-due ${task.dueDate && new Date(task.dueDate).getTime() < now ? 'is-overdue' : ''}`}
                      >
                        {relativeDue(task.dueDate, now)}
                      </span>
                      <span className="dashboard-row-arrow" aria-hidden="true">
                        ↗
                      </span>
                    </Link>
                  ))}
                </div>
              ) : (
                <div className="dashboard-empty">
                  <span aria-hidden="true">✓</span>
                  <strong>All clear for now</strong>
                  <p>
                    No open tasks in this overview. Visit Projects to plan your
                    next step.
                  </p>
                  <Link to="/projects">Browse projects</Link>
                </div>
              )}
            </section>

            <section
              className="dashboard-panel dashboard-progress"
              aria-labelledby="dashboard-progress-title"
            >
              <div className="dashboard-section-heading">
                <div>
                  <p className="dashboard-kicker">WORK DISTRIBUTION</p>
                  <h2 id="dashboard-progress-title">Task progress</h2>
                  <p>Current status of loaded work.</p>
                </div>
              </div>
              <div className="dashboard-progress-main">
                <div
                  className="dashboard-progress-ring"
                  role="img"
                  aria-label={`${completion}% of ${total} loaded tasks completed`}
                  style={{ '--progress': `${completion}%` } as CSSProperties}
                >
                  <div>
                    <strong>{completion}%</strong>
                    <span>completed</span>
                  </div>
                </div>
                <div className="dashboard-progress-copy">
                  <strong>{total}</strong>
                  <span>tasks loaded</span>
                  <small>
                    {done} completed · {total - done} not completed
                  </small>
                </div>
              </div>
              <div
                className="dashboard-distribution"
                role="img"
                aria-label={STATUS.map(
                  (status) => `${status.label}: ${count(status.key)}`,
                ).join(', ')}
              >
                {STATUS.map(
                  (status) =>
                    count(status.key) > 0 && (
                      <span
                        key={status.key}
                        style={{
                          width: `${(count(status.key) / total) * 100}%`,
                          backgroundColor: status.color,
                        }}
                      />
                    ),
                )}
              </div>
              <div className="dashboard-status-list">
                {STATUS.map((status) => (
                  <div key={status.key}>
                    <span
                      className="dashboard-legend-dot"
                      style={{ backgroundColor: status.color }}
                    />
                    <span>{status.label}</span>
                    <strong>{count(status.key)}</strong>
                  </div>
                ))}
              </div>
            </section>
          </div>
          <p className="dashboard-data-note">
            Overview shows up to 50 recently returned tasks, not necessarily
            every task in the workspace. Counts and percentages reflect only
            this sample.
          </p>
        </>
      )}
    </div>
  )
}
