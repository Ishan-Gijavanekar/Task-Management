/** Apply an opt-in theme before React mounts, without an inline CSP exception. */
try {
  if (window.localStorage.getItem('taskflow.theme') === 'dark') {
    document.documentElement.dataset.theme = 'dark'
    document.documentElement.style.colorScheme = 'dark'
  }
} catch {
  // Local storage can be disabled by browser/privacy settings.
}
