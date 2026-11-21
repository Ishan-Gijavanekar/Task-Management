import { EmptyState } from './states'

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
          This module is connected and ready for its collaboration workflow.
        </EmptyState>
      </div>
    </div>
  )
}
