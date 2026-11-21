import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { LoadingScreen } from '../../ui/states'
import { useWorkspace } from './workspace-context'

export function WorkspaceGate({ children }: { children: ReactNode }) {
  const { workspace, loading } = useWorkspace()
  if (loading) return <LoadingScreen label="Opening your workspace" />
  if (workspace) return children
  return (
    <div className="page">
      <section className="workspace-onboarding">
        <span className="onboarding-mark" aria-hidden="true">
          ◇
        </span>
        <p className="eyebrow">Start together</p>
        <h1>Create or open a workspace</h1>
        <p>Workspaces keep projects, tasks, and members in one secure place.</p>
        <Link className="button primary" to="/workspaces">
          Set up a workspace
        </Link>
      </section>
    </div>
  )
}
