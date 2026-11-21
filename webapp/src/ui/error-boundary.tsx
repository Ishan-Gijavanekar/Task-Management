import { Component, type ErrorInfo, type ReactNode } from 'react'

export class ErrorBoundary extends Component<
  { children: ReactNode },
  { failed: boolean }
> {
  override state = { failed: false }

  static getDerivedStateFromError() {
    return { failed: true }
  }

  override componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Application render failed', error, info.componentStack)
  }

  override render() {
    if (this.state.failed) {
      return (
        <main className="centered-state">
          <span className="empty-icon" aria-hidden="true">
            !
          </span>
          <h1>We hit a snag</h1>
          <p>Your work is safe. Reload the page to try again.</p>
          <button
            className="button primary"
            type="button"
            onClick={() => window.location.reload()}
          >
            Reload Taskflow
          </button>
        </main>
      )
    }
    return this.props.children
  }
}
