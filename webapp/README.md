# Taskflow web application

Phases 13 and 14 provide the browser application for the task-management platform. It is a strict TypeScript, React, and Vite application with responsive authentication, workspace, project, and task journeys.

## Phase 14 workflow

- Create or reopen a workspace, switch between locally remembered workspaces, and update workspace settings.
- Create projects and use responsive board or list views with URL-backed search, filters, sorting, saved filters, and cursor pagination.
- Create and edit tasks with assignees, status, priority, dates, labels, checklists, and dependencies; archive and restore tasks from the archived view.
- Move tasks by drag and drop or the accessible move selector. Optimistic changes roll back on failure, and version conflicts offer reload or intentional reapply actions.
- Open task deep links without losing the current filters. Mutation requests use idempotency keys, while server responses remain the cache source of truth.

The current public API does not expose a workspace-list endpoint. The browser therefore remembers only workspace IDs and names that the signed-in user creates or explicitly opens; authorization and membership are still enforced by the backend.

## Run locally

```sh
npm install
npm run dev
```

Vite proxies `/api` to the gateway at `http://localhost:8088`. Set `VITE_API_URL` only when the public API lives at a different origin; include the `/api/v1` prefix.

## Quality commands

```sh
npm run format:check
npm run lint
npm test
npm run build
npx playwright install chromium
npm run test:e2e
```

The production build fails if compressed JavaScript exceeds 180 KiB or CSS exceeds 30 KiB. Playwright exercises authentication plus the primary workspace/project/task workflow and automated accessibility on desktop and mobile Chromium profiles.

## Architecture

- `src/api` contains contract types, centralized error mapping, retry rules, cancellation support, and the cookie-aware API client.
- `src/features` owns authentication, shell, workspace, project, task, member, notification, and settings modules.
- `src/ui` contains accessible shared states, network feedback, branding, and the render error boundary.
- TanStack Query supplies the request cache. Server responses remain authoritative and failed mutations are never retried automatically.
- React Router guards authenticated routes and restores only same-origin relative return paths.

Access tokens exist only in memory. Refresh sessions use the gateway's `HttpOnly`, `SameSite` cookie, API calls include credentials, and logout/login events synchronize across tabs through `BroadcastChannel`. The UI never makes privileged authorization decisions.
