import type { ReactNode } from 'react'

export function LoadingScreen({ label = 'Loading' }: { label?: string }) {
  return (
    <main className="centered-state" aria-busy="true" aria-live="polite">
      <span className="spinner" aria-hidden="true" />
      <p>{label}…</p>
    </main>
  )
}

export function EmptyState({
  title,
  children,
}: {
  title: string
  children: ReactNode
}) {
  return (
    <section className="empty-state">
      <span className="empty-icon" aria-hidden="true">
        ✦
      </span>
      <h2>{title}</h2>
      <p>{children}</p>
    </section>
  )
}
