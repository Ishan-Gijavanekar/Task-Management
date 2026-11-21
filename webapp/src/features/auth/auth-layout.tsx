import { Outlet } from 'react-router-dom'
import { Brand } from '../../ui/brand'

export function AuthLayout() {
  return (
    <div className="auth-layout">
      <aside className="auth-story" aria-label="Taskflow introduction">
        <Brand />
        <div>
          <p className="eyebrow">Clear work. Calm teams.</p>
          <h1>Give every good idea a clear next step.</h1>
          <p>
            Plan projects, bring people together, and keep momentum visible
            without turning work into busywork.
          </p>
        </div>
        <blockquote>“One shared rhythm for the whole team.”</blockquote>
      </aside>
      <main className="auth-main">
        <div className="mobile-brand">
          <Brand />
        </div>
        <Outlet />
      </main>
    </div>
  )
}
