export function safeReturnUrl(value: string | null) {
  if (!value || !value.startsWith('/') || value.startsWith('//')) return '/'
  try {
    const url = new URL(value, window.location.origin)
    return url.origin === window.location.origin
      ? `${url.pathname}${url.search}${url.hash}`
      : '/'
  } catch {
    return '/'
  }
}
