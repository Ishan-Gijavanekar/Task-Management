import type { ErrorEnvelope } from './types'

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly requestId: string | undefined
  readonly fields: Record<string, unknown> | undefined

  constructor(
    status: number,
    code: string,
    message: string,
    requestId?: string,
    fields?: Record<string, unknown>,
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.requestId = requestId
    this.fields = fields
  }
}

export async function mapApiError(response: Response): Promise<ApiError> {
  let body: ErrorEnvelope | undefined
  try {
    body = (await response.json()) as ErrorEnvelope
  } catch {
    // Proxies may return an empty or non-JSON response.
  }
  return new ApiError(
    response.status,
    body?.error.code ?? 'REQUEST_FAILED',
    body?.error.message ?? defaultMessage(response.status),
    body?.error.requestId,
    body?.error.fields,
  )
}

function defaultMessage(status: number) {
  if (status === 401) return 'Your session has expired. Please sign in again.'
  if (status === 403) return 'You do not have permission to do that.'
  if (status === 404) return 'The requested item could not be found.'
  if (status === 429) return 'Too many requests. Please wait and try again.'
  if (status >= 500) return 'The service is temporarily unavailable.'
  return 'The request could not be completed.'
}

export function errorMessage(error: unknown) {
  if (error instanceof ApiError) return error.message
  if (error instanceof TypeError) return 'Check your connection and try again.'
  return 'Something went wrong. Please try again.'
}
