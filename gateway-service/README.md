# API gateway

The API gateway is the single browser-facing HTTP service for the task management application. It exposes the stable `/api/v1` API on port `8088`, validates access tokens before protected requests, forwards refresh cookies, applies public HTTP controls, and routes calls to the user and task services. Domain authorization and business rules remain in those services.

## Run locally

From the repository root, run `docker compose up --build`. The browser API is available at `http://localhost:8088/api/v1`; the OpenAPI contract is available at `http://localhost:8088/openapi.json`. Only the gateway is published to the host by the root Compose stack.

For a standalone process, copy `.env.example`, provide the same JWT signing secret used by the user service, and run:

```sh
go run ./cmd/server
```

`PUBLIC_ORIGIN` must be one explicit absolute origin. Wildcards and comma-separated origin lists are rejected because browser credentials are enabled.

## Contract workflow

The source contract is `api/openapi.json`, and `api/generated.ts` is the committed frontend model output.

```sh
make contract
make generate
make types-check
```

`make contract` compares the current contract to `api/compatibility-baseline.json`. Removing an operation, required response field, or existing enum value fails the check. Update the baseline only as part of an explicitly approved breaking API release. `make types-check` fails when generated frontend types are stale.

## Operations

- `GET /livez` reports process liveness.
- `GET /readyz` verifies both upstream services within the configured timeout.
- `GET /metrics` exposes a Prometheus text counter.
- `GET /openapi.json` serves the exact embedded public contract.

Every response includes `X-Request-ID`. Incoming W3C `traceparent`, `tracestate`, and `baggage` headers are continued and propagated upstream. Set `OTEL_EXPORTER_OTLP_ENDPOINT` to export spans over OTLP/gRPC.

See [API policy](docs/api-policy.md) for compatibility, errors, retries, idempotency, and deprecation rules.
