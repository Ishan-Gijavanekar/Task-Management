# Task service

The task service owns workspaces, projects, memberships, invitations, workspace-scoped tasks, ordering, dependencies, immutable activity and audit records, idempotency, and optimistic concurrency. It exposes REST at `/api/v1` and the versioned `task.v1.TaskService` and `task.v1.WorkspaceService` gRPC contracts.

Run `docker compose up --build` in this directory. MongoDB runs as a single-node replica set because mutations use transactions to store task state, activity, and idempotency results together. Use `GET /livez` for process health, `GET /readyz` for MongoDB readiness, and `GET /metrics` for Prometheus-format metrics.

All endpoints require an access token signed by the user service. A user creates a workspace through `POST /api/v1/workspaces` and becomes its owner. Owners and admins can create projects and manage invitations and members; only owners can archive a workspace or transfer ownership. Invitation tokens are returned only when created or resent, while MongoDB stores only their SHA-256 hashes. Invitations expire after seven days by default, are single-use, can be revoked, and are limited to ten sends per administrator per workspace per hour.

Authorization is enforced inside services and documented in [docs/authorization.md](docs/authorization.md). Membership lookups use a bounded 30-second cache that is explicitly invalidated after acceptance, role changes, suspension, removal, and ownership transfer. Private projects and their tasks are limited to owners and admins. Removed users lose access immediately; historical task attribution is preserved and recorded in the administrative audit event.

List endpoints use opaque cursors and stable `(position, _id)` ordering. Mutations accept `Idempotency-Key`; updates, archive, restore, reorder, and batches also require the current `version`. A stale version returns HTTP 409 or gRPC `ABORTED`.

The service creates its indexes at startup. Phase 11 adds `workspaces`, `projects`, `workspace_invitations`, `workspace_audit_events`, and `workspace_rate_limits`, and expands `workspace_memberships` with status and timestamps. Existing memberships must be backfilled with `status: "active"`, `createdAt`, and `updatedAt` before enabling the Phase 11 binary. Production deployments should create the committed indexes in a deployment job before raising traffic.

Rollback is schema-safe: the Phase 10 binary ignores the new collections and added membership fields. Do not delete Phase 11 data during rollback. Ownership, membership, and invitation mutations made while Phase 11 was active remain authoritative when it is redeployed.
