import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { Brand } from '../../ui/brand'
import { NetworkStatus } from '../../ui/network-status'
import { useAuth } from '../auth/auth-context'

const navItems = [
  ['/', 'Overview', '⌂'],
  ['/workspaces', 'Workspaces', '◇'],
  ['/projects', 'Projects', '▦'],
  ['/tasks', 'Tasks', '✓'],
  ['/members', 'Members', '♧'],
  ['/notifications', 'Notifications', '◌'],
] as const

export function AppShell() {
  const { user, logout } = useAuth()
  const [open, setOpen] = useState(false)

  return (
    <div className="app-shell">
      <NetworkStatus />
      <aside className={`sidebar ${open ? 'open' : ''}`}>
        <div className="sidebar-top">
          <Brand />
          <button
            className="icon-button close-menu"
            type="button"
            onClick={() => setOpen(false)}
            aria-label="Close navigation"
          >
            ×
          </button>
        </div>
        <div className="workspace-switcher">
          <span className="workspace-avatar">A</span>
          <span>
            <small>Workspace</small>
            <strong>Acme Studio</strong>
          </span>
          <span aria-hidden="true">⌄</span>
        </div>
        <nav aria-label="Primary navigation">
          {navItems.map(([to, label, icon]) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              onClick={() => setOpen(false)}
            >
              <span aria-hidden="true">{icon}</span>
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <NavLink to="/settings">
            <span aria-hidden="true">⚙</span>Settings
          </NavLink>
          <div className="profile-block">
            <span className="profile-avatar">
              {user?.name.slice(0, 2).toUpperCase()}
            </span>
            <span>
              <strong>{user?.name}</strong>
              <small>{user?.email}</small>
            </span>
            <button
              className="icon-button"
              type="button"
              onClick={() => void logout()}
              aria-label="Sign out"
            >
              ↪
            </button>
          </div>
        </div>
      </aside>
      {open && (
        <button
          className="scrim"
          type="button"
          onClick={() => setOpen(false)}
          aria-label="Close navigation"
        />
      )}
      <div className="shell-content">
        <header className="mobile-header">
          <button
            className="icon-button"
            type="button"
            onClick={() => setOpen(true)}
            aria-label="Open navigation"
          >
            ☰
          </button>
          <Brand />
          <span className="profile-avatar small">
            {user?.name.slice(0, 2).toUpperCase()}
          </span>
        </header>
        <Outlet />
      </div>
    </div>
  )
}
