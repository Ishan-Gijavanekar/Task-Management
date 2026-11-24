import type { ReactNode } from 'react'

export function LoadingScreen({ label = 'Loading' }: { label?: string }) {
  return (
    <div
      className="centered-state"
      role="status"
      aria-busy="true"
      aria-live="polite"
    >
      <span className="spinner" aria-hidden="true" />
      <p>{label}…</p>
    </div>
  )
}

export function EmptyState({
  title,
  children,
  action,
}: {
  title: string
  children: ReactNode
  action?: ReactNode
}) {
  return (
    <section className="empty-state">
      <svg
        className="empty-artwork"
        width="126"
        height="92"
        viewBox="0 0 126 92"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        aria-hidden="true"
      >
        <rect x="21" y="10" width="84" height="64" rx="7" />
        <path d="M37 31h54M37 43h54M37 55h28" />
        <circle cx="96" cy="68" r="16" fill="#FFFFFF" />
        <path d="m107 79 12 11M89 68h14" />
      </svg>
      <h2>{title}</h2>
      <p>{children}</p>
      {action}
    </section>
  )
}
