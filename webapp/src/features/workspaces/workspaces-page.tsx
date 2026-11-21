import { EmptyState } from '../../ui/states'

export function WorkspacesPage() {
  return (
    <ModulePage
      eyebrow="Organization"
      title="Workspaces"
      description="Your shared homes for projects and people."
    />
  )
}

export function ModulePage({
  eyebrow,
  title,
  description,
}: {
  eyebrow: string
  title: string
  description: string
}) {
  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">{eyebrow}</p>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>
      </header>
      <div className="panel module-panel">
        <EmptyState title={`${title} are ready`}>
          This foundation is connected and ready for the Phase 14
          task-management experience.
        </EmptyState>
      </div>
    </div>
  )
}
