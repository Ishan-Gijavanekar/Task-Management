import { useState, type ReactNode } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { Brand } from '../../ui/brand'
import { NetworkStatus } from '../../ui/network-status'
import { useAuth } from '../auth/auth-context'
import {
  WorkspaceProvider,
  useWorkspace,
} from '../workspaces/workspace-context'
import './app-shell.css'

type IconName =
  | 'home'
  | 'layers'
  | 'grid'
  | 'check'
  | 'people'
  | 'bell'
  | 'settings'
  | 'logout'
  | 'menu'
  | 'close'
  | 'chevron'

function NavIcon({ name }: { name: IconName }) {
  const paths: Record<IconName, ReactNode> = {
    home: (
      <>
        <path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1Z" />
        <path d="M9 21v-7h6v7" />
      </>
    ),
    layers: (
      <>
        <rect x="3" y="3" width="18" height="18" rx="3" />
        <path d="M3 9h18M9 9v12" />
      </>
    ),
    grid: (
      <>
        <rect x="3" y="3" width="7" height="7" rx="1" />
        <rect x="14" y="3" width="7" height="7" rx="1" />
        <rect x="3" y="14" width="7" height="7" rx="1" />
        <rect x="14" y="14" width="7" height="7" rx="1" />
      </>
    ),
    check: (
      <>
        <rect x="3" y="3" width="18" height="18" rx="3" />
        <path d="m8 12 3 3 5-6" />
      </>
    ),
    people: (
      <>
        <circle cx="9" cy="8" r="3" />
        <path d="M3 20v-2a6 6 0 0 1 12 0v2M17 5a3 3 0 0 1 0 6M18 15a5 5 0 0 1 3 5" />
      </>
    ),
    bell: (
      <>
        <path d="M18 8a6 6 0 1 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4" />
      </>
    ),
    settings: (
      <>
        <circle cx="12" cy="12" r="3" />
        <path
          d="m19.4 15 1.4 1.1-2 3.5-1.7-.6a9 9 0 0 1-2 1.2l-.3 1.8h-4l-.3-1.8a9 9 0 0 1-2-1.2l-1.7.6-2-3.5L6.2 15a9 9 0 0 1 0-2l-1.4-1.1 2-3.5 1.7.6a9 9 0 0 1 2-1.2L10.8 6h4l.3 1.8a9 9 0 0 1 2 1.2l1.7-.6 2 3.5-1.4 1.1a9 9 0 0 1 0 2Z"
          transform="translate(0 -2)"
        />
      </>
    ),
    logout: (
      <>
        <path d="M10 4H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h5M15 8l4 4-4 4M9 12h10" />
      </>
    ),
    menu: <path d="M4 6h16M4 12h16M4 18h16" />,
    close: <path d="M5 5l14 14M19 5 5 19" />,
    chevron: <path d="m7 10 5 5 5-5" />,
  }
  return (
    <svg
      width="19"
      height="19"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.65"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {paths[name]}
    </svg>
  )
}

const navigation: { to: string; label: string; icon: IconName }[] = [
  { to: '/', label: 'Overview', icon: 'home' },
  { to: '/workspaces', label: 'Workspaces', icon: 'layers' },
  { to: '/projects', label: 'Projects', icon: 'grid' },
  { to: '/tasks', label: 'Tasks', icon: 'check' },
  { to: '/members', label: 'Members', icon: 'people' },
  { to: '/notifications', label: 'Notifications', icon: 'bell' },
]

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
  const initial = user?.name?.trim().slice(0, 2).toUpperCase() || '?'

  return (
    <div className="app-shell shell-refresh">
      <NetworkStatus />
      <aside
        className={`sidebar ${open ? 'open' : ''}`}
        aria-label="Application sidebar"
      >
        <div className="sidebar-top">
          <Brand />
          <button
            className="icon-button close-menu"
            type="button"
            onClick={() => setOpen(false)}
            aria-label="Close navigation"
          >
            <NavIcon name="close" />
          </button>
        </div>
        <div className="shell-sidebar-section-label">CURRENT WORKSPACE</div>
        <NavLink
          className="workspace-switcher"
          to="/workspaces"
          onClick={() => setOpen(false)}
        >
          <span className="workspace-avatar" aria-hidden="true">
            {workspace?.name.slice(0, 1).toUpperCase() ?? '+'}
          </span>
          <span>
            <small>Workspace</small>
            <strong>{workspace?.name ?? 'Choose workspace'}</strong>
          </span>
          <NavIcon name="chevron" />
        </NavLink>
        <div className="shell-sidebar-section-label shell-nav-caption">
          WORKSPACE
        </div>
        <nav aria-label="Primary navigation">
          {navigation.map(({ to, label, icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              onClick={() => setOpen(false)}
            >
              <NavIcon name={icon} />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <NavLink to="/settings" onClick={() => setOpen(false)}>
            <NavIcon name="settings" />
            <span>Settings</span>
          </NavLink>
          <div className="profile-block">
            <span className="profile-avatar" aria-hidden="true">
              {initial}
            </span>
            <span>
              <strong>{user?.name || 'Your account'}</strong>
              <small>{user?.email}</small>
            </span>
            <button
              className="icon-button"
              type="button"
              onClick={() => void logout()}
              aria-label="Sign out"
              title="Sign out"
            >
              <NavIcon name="logout" />
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
            <NavIcon name="menu" />
          </button>
          <Brand />
          <span className="profile-avatar small" aria-label="Signed-in user">
            {initial}
          </span>
        </header>
        <main id="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
