import { ThemeToggle } from '../../ui/theme-toggle'
import { useEffect, useState, type ReactNode } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
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
  | 'projects'
  | 'tasks'
  | 'people'
  | 'bell'
  | 'settings'
  | 'logout'
  | 'more'
  | 'close'
  | 'chevron'
const strokes: Record<IconName, ReactNode> = {
  home: (
    <>
      <path d="M3 10.5 12 3l9 7.5V21H3V10.5Z" />
      <path d="M9 21v-7h6v7" />
    </>
  ),
  projects: (
    <>
      <rect x="3" y="3" width="8" height="8" rx="1.5" />
      <rect x="13" y="3" width="8" height="8" rx="1.5" />
      <rect x="3" y="13" width="8" height="8" rx="1.5" />
      <rect x="13" y="13" width="8" height="8" rx="1.5" />
    </>
  ),
  tasks: (
    <>
      <rect x="3" y="4" width="18" height="17" rx="2" />
      <path d="m8 12 2.5 2.5L16 9" />
    </>
  ),
  people: (
    <>
      <circle cx="9" cy="8" r="3" />
      <path d="M3 21v-2a6 6 0 0 1 12 0v2M17 5a3 3 0 0 1 0 6M18 15a5 5 0 0 1 3 5" />
    </>
  ),
  bell: (
    <>
      <path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9" />
      <path d="M10 21h4" />
    </>
  ),
  settings: (
    <>
      <circle cx="12" cy="12" r="3" />
      <path d="M19 12a7 7 0 0 0-.1-1l2-1.5-2-3.5-2.3 1a8 8 0 0 0-1.7-1L14.5 3h-5L9 6a8 8 0 0 0-1.7 1L5 6 3 9.5 5 11a7 7 0 0 0 0 2l-2 1.5L5 18l2.3-1a8 8 0 0 0 1.7 1l.5 3h5l.4-3a8 8 0 0 0 1.7-1l2.3 1 2-3.5-2-1.5c.1-.3.1-.7.1-1Z" />
    </>
  ),
  logout: (
    <>
      <path d="M10 4H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h5" />
      <path d="m15 8 4 4-4 4M9 12h10" />
    </>
  ),
  more: (
    <>
      <circle cx="5" cy="12" r="1" />
      <circle cx="12" cy="12" r="1" />
      <circle cx="19" cy="12" r="1" />
    </>
  ),
  close: <path d="M5 5 19 19M19 5 5 19" />,
  chevron: <path d="m7 10 5 5 5-5" />,
}
function NavIcon({ name }: { name: IconName }) {
  return (
    <svg
      width="19"
      height="19"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {strokes[name]}
    </svg>
  )
}
const navigation: { to: string; label: string; icon: IconName }[] = [
  { to: '/', label: 'Overview', icon: 'home' },
  { to: '/projects', label: 'Projects', icon: 'projects' },
  { to: '/tasks', label: 'Tasks', icon: 'tasks' },
  { to: '/members', label: 'Members', icon: 'people' },
  { to: '/notifications', label: 'Inbox', icon: 'bell' },
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
  const location = useLocation()
  const [moreMenuPath, setMoreMenuPath] = useState<string | null>(null)
  const moreOpen = moreMenuPath === location.pathname
  useEffect(() => {
    function onEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') setMoreMenuPath(null)
    }
    window.addEventListener('keydown', onEscape)
    return () => window.removeEventListener('keydown', onEscape)
  }, [])
  const initial = user?.name?.trim().slice(0, 2).toUpperCase() || '?'
  return (
    <div className="app-shell shell-refresh">
      <a href="#main-content" className="shell-skip-link">
        Skip to content
      </a>
      <NetworkStatus />
      <aside className="sidebar" aria-label="Application sidebar">
        <div className="sidebar-top">
          <Brand />
        </div>
        <div className="shell-sidebar-section-label">YOUR SPACE</div>
        <NavLink
          className="workspace-switcher"
          to="/workspaces"
          title={workspace?.name ?? 'Switch workspace'}
        >
          <span className="workspace-avatar" aria-hidden="true">
            {workspace?.name.slice(0, 1).toUpperCase() ?? '+'}
          </span>
          <span className="workspace-name">
            <small>Workspace</small>
            <strong>{workspace?.name ?? 'Choose workspace'}</strong>
          </span>
          <NavIcon name="chevron" />
        </NavLink>
        <div className="shell-sidebar-section-label shell-nav-caption">
          WORKSPACE
        </div>
        <nav aria-label="Primary navigation">
          {navigation.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              title={item.label}
            >
              <NavIcon name={item.icon} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <ThemeToggle />
          <NavLink to="/settings" title="Settings">
            <NavIcon name="settings" />
            <span>Settings</span>
          </NavLink>
          <div className="profile-block">
            <span className="profile-avatar" aria-hidden="true">
              {initial}
            </span>
            <span className="profile-copy">
              <strong>{user?.name || 'Your account'}</strong>
              <small>{user?.email}</small>
            </span>
            <button
              type="button"
              className="icon-button"
              onClick={() => void logout()}
              title="Sign out"
              aria-label="Sign out"
            >
              <NavIcon name="logout" />
            </button>
          </div>
        </div>
      </aside>
      <div className="shell-content">
        <main id="main-content" tabIndex={-1}>
          <Outlet />
        </main>
      </div>
      <nav className="mobile-tabs" aria-label="Mobile navigation">
        {navigation.slice(0, 3).map((item) => (
          <NavLink key={item.to} to={item.to} end={item.to === '/'}>
            <NavIcon name={item.icon} />
            <span>{item.label}</span>
          </NavLink>
        ))}
        <button
          type="button"
          className={moreOpen ? 'active' : ''}
          onClick={() =>
            setMoreMenuPath((current) =>
              current === location.pathname ? null : location.pathname,
            )
          }
          aria-expanded={moreOpen}
          aria-controls="mobile-more-menu"
        >
          <NavIcon name="more" />
          <span>More</span>
        </button>
      </nav>
      {moreOpen && (
        <>
          <button
            className="mobile-more-scrim"
            type="button"
            aria-label="Close more menu"
            onClick={() => setMoreMenuPath(null)}
          />
          <section
            className="mobile-more-sheet"
            id="mobile-more-menu"
            aria-label="More destinations"
          >
            <div className="mobile-more-heading">
              <strong>More</strong>
              <button
                className="icon-button"
                type="button"
                aria-label="Close menu"
                onClick={() => setMoreMenuPath(null)}
              >
                <NavIcon name="close" />
              </button>
            </div>
            {[
              {
                to: '/workspaces',
                label: 'Workspaces',
                icon: 'projects' as IconName,
              },
              ...navigation.slice(3),
              {
                to: '/settings',
                label: 'Settings',
                icon: 'settings' as IconName,
              },
            ].map((item) => (
              <NavLink key={item.to} to={item.to}>
                <NavIcon name={item.icon} />
                {item.label}
              </NavLink>
            ))}
            <ThemeToggle compact />
            <button
              type="button"
              className="mobile-signout"
              onClick={() => void logout()}
            >
              <NavIcon name="logout" />
              Sign out
            </button>
          </section>
        </>
      )}
    </div>
  )
}
