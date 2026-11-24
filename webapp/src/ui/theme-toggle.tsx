import { useLayoutEffect, useState } from 'react'
import './theme-toggle.css'

type Theme = 'light' | 'dark'
const STORAGE_KEY = 'taskflow.theme'

function initialTheme(): Theme {
  try {
    return window.localStorage.getItem(STORAGE_KEY) === 'dark'
      ? 'dark'
      : 'light'
  } catch {
    return 'light'
  }
}

/** Both desktop and mobile toggles stay in sync using the document event. */
export function ThemeToggle({ compact = false }: { compact?: boolean }) {
  const [theme, setTheme] = useState<Theme>(initialTheme)

  useLayoutEffect(() => {
    document.documentElement.dataset.theme = theme
    document.documentElement.style.colorScheme = theme
    try {
      window.localStorage.setItem(STORAGE_KEY, theme)
    } catch {
      /* Private browsing */
    }
    // Multiple visible toggles within the same app stay in sync.
    const onChange = (event: Event) => {
      const next = (event as CustomEvent<Theme>).detail
      if (next !== theme) setTheme(next)
    }
    const onStorage = (event: StorageEvent) => {
      if (event.key === STORAGE_KEY)
        setTheme(event.newValue === 'dark' ? 'dark' : 'light')
    }
    window.addEventListener('taskflow-theme-change', onChange)
    window.addEventListener('storage', onStorage)
    return () => {
      window.removeEventListener('taskflow-theme-change', onChange)
      window.removeEventListener('storage', onStorage)
    }
  }, [theme])

  function toggle() {
    const next: Theme = theme === 'light' ? 'dark' : 'light'
    setTheme(next)
    window.dispatchEvent(
      new CustomEvent<Theme>('taskflow-theme-change', { detail: next }),
    )
  }

  const isDark = theme === 'dark'
  return (
    <button
      type="button"
      className={`theme-toggle${compact ? ' theme-toggle--compact' : ''}`}
      onClick={toggle}
      aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      aria-pressed={isDark}
      title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      <svg
        viewBox="0 0 24 24"
        width="18"
        height="18"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        {isDark ? (
          <>
            <circle cx="12" cy="12" r="4" />
            <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
          </>
        ) : (
          <path d="M20.5 14A8.5 8.5 0 0 1 10 3.5 8.5 8.5 0 1 0 20.5 14Z" />
        )}
      </svg>
      <span className="theme-toggle__label">
        {isDark ? 'Light mode' : 'Dark mode'}
      </span>
      {!compact && (
        <span className="theme-toggle__switch" aria-hidden="true">
          <i />
        </span>
      )}
    </button>
  )
}
