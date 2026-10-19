# Task service

The task service owns workspace-scoped tasks, ordering, dependencies, immutable activity records, idempotency, and optimistic concurrency. It exposes REST at `/api/v1` and the versioned `task.v1.TaskService` gRPC contract.

Run `docker compose up --build` in this directory. MongoDB runs as a single-node replica set because mutations use transactions to store task state, activity, and idempotency results together. Use `GET /livez` for process health, `GET /readyz` for MongoDB readiness, and `GET /metrics` for Prometheus-format metrics.

All task endpoints require an access token signed by the user service. Workspace membership is checked inside the service for every operation. Phase 11 will add membership-management APIs; until then, seed `workspace_memberships` documents containing `workspaceId`, `userId`, and `role`.

List endpoints use opaque cursors and stable `(position, _id)` ordering. Mutations accept `Idempotency-Key`; updates, archive, restore, reorder, and batches also require the current `version`. A stale version returns HTTP 409 or gRPC `ABORTED`.

The service creates its indexes at startup. Production deployments should move index creation to a deployment job before raising traffic. Rollback is safe because this phase adds new collections and does not modify user-service data.
