import { describe, expect, it } from 'vitest'
import { safeReturnUrl } from './safe-return'

describe('safeReturnUrl', () => {
  it('keeps local paths with query strings and fragments', () => {
    expect(safeReturnUrl('/tasks?status=todo#today')).toBe(
      '/tasks?status=todo#today',
    )
  })

  it.each([
    'https://evil.example',
    '//evil.example/path',
    'javascript:alert(1)',
    null,
  ])('rejects unsafe return target %s', (target) =>
    expect(safeReturnUrl(target)).toBe('/'),
  )
})
