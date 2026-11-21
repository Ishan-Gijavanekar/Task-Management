import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { Brand } from '../../ui/brand'
import { NetworkStatus } from '../../ui/network-status'
import { useAuth } from '../auth/auth-context'
import {
  WorkspaceProvider,
  useWorkspace,
} from '../workspaces/workspace-context'

const navItems = [
  ['/', 'Overview', '⌂'],
  ['/workspaces', 'Workspaces', '◇'],
  ['/projects', 'Projects', '▦'],
  ['/tasks', 'Tasks', '✓'],
  ['/members', 'Members', '♧'],
  ['/notifications', 'Notifications', '◌'],
] as const

export function AppShell() {
  return (
    <WorkspaceProvider>
      <Shell />
    </WorkspaceProvider>
  )
}

function Shell() {
  const { user, logout } = useAuth()
  const { workspace } = useWorkspace()
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
        <NavLink
          className="workspace-switcher"
          to="/workspaces"
          onClick={() => setOpen(false)}
        >
          <span className="workspace-avatar">
            {workspace?.name.slice(0, 1).toUpperCase() ?? '+'}
          </span>
          <span>
            <small>Workspace</small>
            <strong>{workspace?.name ?? 'Choose workspace'}</strong>
          </span>
          <span aria-hidden="true">⌄</span>
        </NavLink>
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
