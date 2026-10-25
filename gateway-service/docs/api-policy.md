# Public API policy

## Stable API surface

The supported browser API is the gateway origin under `/api/v1`. Browser clients must not call service containers directly. Paths outside the gateway allowlist return `404`, and unsupported methods are never forwarded.

Within `/api/v1`, additions are allowed when existing clients continue to work. Compatible changes include new optional fields, endpoints, response headers, and enum values only where clients have been documented to tolerate unknown values. The following are breaking changes and require a new major API prefix or explicit migration approval:

- removing or renaming an operation, field, or accepted value;
- changing a field type or making an optional request field required;
- removing a required response field or an established enum value;
- changing authentication, authorization, status-code, pagination, or idempotency semantics in a way that breaks an existing client.

The compatibility baseline and generated frontend types are checked in CI. Contract changes must include their generated type update.

## Deprecation

A deprecated operation or field is marked `deprecated: true` in OpenAPI and documented with its replacement. The gateway also emits `Deprecation: true`, `Sunset`, and a `Link` header to migration guidance when an entire endpoint is being retired. Supported `/api/v1` behavior receives at least 90 days of notice before removal, except when an urgent security response requires faster action. Removal then occurs in a new major API version unless consumers have explicitly completed the migration.

## Authentication and cookies

The gateway is the browser authentication boundary. It accepts credentials only on documented authentication routes, forwards the refresh session as an `HttpOnly` cookie, and validates bearer access tokens before protected calls. Production refresh cookies must use `Secure` and an appropriate `SameSite` policy. Access tokens belong in memory and must not be persisted in browser local storage.

The gateway validation is an early rejection layer. User and task services remain responsible for domain authorization, workspace isolation, account status, role checks, and mutation rules.

## Errors and request IDs

Errors use one envelope across gateway-generated and normalized upstream failures:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "request is invalid",
    "requestId": "01JABC...",
    "fields": {
      "title": "is required"
    }
  }
}
```

`code` is stable and intended for program logic. `message` is safe to display but may change. `fields` is optional and contains safe validation messages; it never exposes stack traces, database errors, tokens, or internal identifiers. Clients should send their `X-Request-ID` when reporting an error. The gateway accepts a safe caller ID or generates one and returns it on every response.

## Pagination and limits

Task lists use an opaque `cursor` and return `nextCursor` when another page exists. Clients must pass the cursor back unchanged and must restart pagination when filters change. User lists currently use `page` and `limit`. Each operation's limits and defaults are defined in OpenAPI.

The gateway limits request bodies to 2 MiB and upstream responses to 8 MiB by default. It applies endpoint-specific fixed-window limits and returns `429` with `Retry-After`. Deployment infrastructure may add stricter limits.

## Cancellation, timeouts, and retries

The gateway cancels its upstream request when the client disconnects or the configured upstream timeout expires. Clients should use their own cancellation signal and a deadline longer than the gateway timeout.

Clients may retry `GET` after connection failures, `429`, `502`, or `503` with exponential backoff and jitter. Honor `Retry-After` when present. Do not retry validation, authentication, authorization, conflict, or not-found responses.

The gateway does not automatically retry mutations. Task mutation endpoints that document `Idempotency-Key` may be retried using the same key and byte-for-byte equivalent logical payload. A key reused with different input is rejected. Generate one high-entropy key per intended mutation and retain it until the final response. Mutations without documented idempotency support require the client to reconcile state with a read before attempting again.
