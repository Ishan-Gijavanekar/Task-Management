import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../api/client'
import { useAuth } from '../auth/auth-context'
import { useWorkspace } from '../workspaces/workspace-context'

export function DashboardPage() {
  const { user } = useAuth()
  const { workspace } = useWorkspace()
  const [now] = useState(Date.now)
  const firstName = user?.name.split(' ')[0] || 'there'
  const projects = useQuery({
    queryKey: ['projects', workspace!.id],
    queryFn: ({ signal }) => api.listProjects(workspace!.id, signal),
  })
  const tasks = useQuery({
    queryKey: ['tasks', workspace!.id, 'overview'],
    queryFn: ({ signal }) =>
      api.listTasks(
        workspace!.id,
        new URLSearchParams({ limit: '50' }),
        signal,
      ),
  })
  const today = new Intl.DateTimeFormat(undefined, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  }).format(new Date())
  const items = tasks.data?.tasks ?? []
  const done = items.filter((task) => task.status === 'done').length
  const active = items.filter((task) =>
    ['in_progress', 'blocked'].includes(task.status),
  ).length
  const dueSoon = items.filter(
    (task) =>
      task.dueDate &&
      new Date(task.dueDate).getTime() < now + 3 * 86_400_000 &&
      task.status !== 'done',
  ).length
  const completion = items.length ? Math.round((done / items.length) * 100) : 0

  return (
    <div className="page dashboard-page">
      <header className="page-header welcome-header">
        <div>
          <p className="eyebrow">{today}</p>
          <h1>Good morning, {firstName}.</h1>
          <p>Here’s the shape of {workspace!.name} today.</p>
        </div>
        <Link className="button primary" to="/tasks">
          View tasks
        </Link>
      </header>
      <section className="metric-grid" aria-label="Work overview">
        <article className="metric-card featured">
          <div>
            <span className="metric-icon coral" aria-hidden="true">
              ↗
            </span>
            <span>Completed</span>
          </div>
          <strong>{done}</strong>
          <p>of {items.length} loaded tasks</p>
          <div className="mini-bars" aria-hidden="true">
            {[35, 52, 45, 72, 58, 88, 68].map((height, index) => (
              <i key={index} style={{ height: `${height}%` }} />
            ))}
          </div>
        </article>
        <article className="metric-card">
          <div>
            <span className="metric-icon blue" aria-hidden="true">
              ◎
            </span>
            <span>In motion</span>
          </div>
          <strong>{active}</strong>
          <p>active or blocked</p>
        </article>
        <article className="metric-card">
          <div>
            <span className="metric-icon amber" aria-hidden="true">
              ◷
            </span>
            <span>Due soon</span>
          </div>
          <strong>{dueSoon}</strong>
          <p>in the next 3 days</p>
        </article>
      </section>
      <div className="dashboard-grid">
        <section className="panel" aria-labelledby="focus-heading">
          <div className="panel-heading">
            <div>
              <p className="eyebrow">Your focus</p>
              <h2 id="focus-heading">Work in motion</h2>
            </div>
            <Link to="/tasks">View all</Link>
          </div>
          <div className="work-list">
            {items.slice(0, 4).map((task) => (
              <article className="work-row" key={task.id}>
                <span className="project-mark coral" aria-hidden="true" />
                <div>
                  <h3>{task.title}</h3>
                  <p>
                    {projects.data?.projects.find(
                      (project) => project.id === task.projectId,
                    )?.name ?? 'Project'}
                  </p>
                </div>
                <span className={`status-chip ${task.status}`}>
                  {task.status.replace('_', ' ')}
                </span>
                <Link
                  className="icon-button"
                  aria-label={`Open ${task.title}`}
                  to={`/tasks?task=${task.id}`}
                >
                  →
                </Link>
              </article>
            ))}
            {!items.length && (
              <p className="muted">
                No tasks yet. Create a project, then add the first task.
              </p>
            )}
          </div>
        </section>
        <section className="panel momentum" aria-labelledby="momentum-heading">
          <div className="panel-heading">
            <div>
              <p className="eyebrow">Team pulse</p>
              <h2 id="momentum-heading">Momentum</h2>
            </div>
            <span className="live-dot">Live</span>
          </div>
          <div
            className="donut"
            style={{
              background: `conic-gradient(var(--brand) 0 ${completion}%, #edf0ed ${completion}%)`,
            }}
            aria-label={`${completion} percent of loaded work complete`}
          >
            <span>
              {completion}%<small>complete</small>
            </span>
          </div>
          <div className="legend">
            <span>
              <i className="coral" />
              Done {done}
            </span>
            <span>
              <i className="blue" />
              Active {active}
            </span>
            <span>
              <i className="muted" />
              Total {items.length}
            </span>
          </div>
        </section>
      </div>
    </div>
  )
}
