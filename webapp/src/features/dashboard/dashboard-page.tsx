import { Link } from 'react-router-dom'
import { useAuth } from '../auth/auth-context'

const work = [
  {
    title: 'Website refresh',
    project: 'Marketing',
    status: 'In progress',
    tone: 'coral',
  },
  {
    title: 'Mobile onboarding',
    project: 'Product',
    status: 'Review',
    tone: 'violet',
  },
  { title: 'Q4 planning', project: 'Operations', status: 'Todo', tone: 'blue' },
]

export function DashboardPage() {
  const { user } = useAuth()
  const firstName = user?.name.split(' ')[0] || 'there'
  const today = new Intl.DateTimeFormat(undefined, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  }).format(new Date())

  return (
    <div className="page dashboard-page">
      <header className="page-header welcome-header">
        <div>
          <p className="eyebrow">{today}</p>
          <h1>Good morning, {firstName}.</h1>
          <p>Here’s the shape of your team’s work today.</p>
        </div>
        <Link className="button primary" to="/tasks/new">
          <span aria-hidden="true">＋</span> New task
        </Link>
      </header>

      <section className="metric-grid" aria-label="Work overview">
        <article className="metric-card featured">
          <div>
            <span className="metric-icon coral" aria-hidden="true">
              ↗
            </span>
            <span>This week</span>
          </div>
          <strong>24</strong>
          <p>
            tasks completed <span className="positive">+18%</span>
          </p>
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
            <span>In progress</span>
          </div>
          <strong>8</strong>
          <p>across 3 projects</p>
        </article>
        <article className="metric-card">
          <div>
            <span className="metric-icon amber" aria-hidden="true">
              ◷
            </span>
            <span>Due soon</span>
          </div>
          <strong>5</strong>
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
            {work.map((item) => (
              <article className="work-row" key={item.title}>
                <span
                  className={`project-mark ${item.tone}`}
                  aria-hidden="true"
                />
                <div>
                  <h3>{item.title}</h3>
                  <p>{item.project}</p>
                </div>
                <span className={`status-chip ${item.tone}`}>
                  {item.status}
                </span>
                <button
                  className="icon-button"
                  type="button"
                  aria-label={`More options for ${item.title}`}
                >
                  •••
                </button>
              </article>
            ))}
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
            aria-label="74 percent of weekly work complete"
          >
            <span>
              74%<small>complete</small>
            </span>
          </div>
          <div className="legend">
            <span>
              <i className="coral" />
              Done 24
            </span>
            <span>
              <i className="blue" />
              Active 8
            </span>
            <span>
              <i className="muted" />
              Planned 6
            </span>
          </div>
        </section>
      </div>
    </div>
  )
}
