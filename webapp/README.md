# Task Management Web Application

The web application is the browser interface for Task Management. Built with React, TypeScript and Vite, it provides account, workspace, project and task workflows through the versioned API Gateway. Data fetching and cache management use TanStack Query; React Router handles application navigation.

## Features

- Registration, login, recovery and authenticated navigation.
- Workspace creation, reopening, switching and settings.
- Project creation, visibility and workflow-oriented task views.
- Responsive board and list layouts with search, filters, sorting and saved filters.
- Task creation and updates with priority, assignee, dates, labels, checklists and dependencies.
- Drag-and-drop task movement with an accessible alternative control.
- Task detail links, comments, archive and restore operations.
- Member management and notification flows.
- Optimistic interactions with rollback and explicit version-conflict handling.

The frontend is a client of the backend API; it does not make authoritative access-control decisions.

## Technology stack

| Area | Technology |
| --- | --- |
| UI | React 19 |
| Language | TypeScript |
| Build tool | Vite 8 |
| Routing | React Router 7 |
| Server state | TanStack Query 5 |
| Unit/component testing | Vitest and Testing Library |
| Browser E2E testing | Playwright |
| Accessibility testing | axe-core / Playwright |
| Formatting and linting | Prettier and ESLint |

## Local development

Start the backend from the repository root:

```bash
docker compose up -d --build
```

Then start the web app:

```bash
cd webapp
npm ci
npm run dev
```

Visit **http://localhost:5173**. During development, Vite proxies `/api` calls to the API Gateway at `http://localhost:8088`. If the gateway uses a different host/origin, set `VITE_API_URL` to its complete public API base URL, including `/api/v1`.

## Source organization

```text
webapp/
├── src/
│   ├── api/          # API client, contract types, error handling and retries
│   ├── features/     # Authentication, workspaces, projects, tasks and settings
│   └── ui/           # Reusable UI and shared feedback components
├── scripts/          # Build-size budget checks
├── package.json
└── README.md
```

The browser keeps access tokens in memory. Refresh sessions use an `HttpOnly`, `SameSite` cookie, and API calls include credentials. Authentication-related events synchronize across tabs through `BroadcastChannel`.

## Workspace and task state

The current public API does not expose a workspace-list endpoint. The browser therefore remembers workspace IDs and names for workspaces the user has created or explicitly opened. Backend membership and authorization checks still apply on every protected request.

Board and list view state uses URL-backed search, filtering and sorting. TanStack Query manages server state; optimistic updates are reverted on failure, and stale entity versions present explicit reload or intentional reapply options. Mutations use idempotency keys where required by the API.

## Development commands

Run from `webapp/`:

| Command | Purpose |
| --- | --- |
| `npm run dev` | Start Vite development server |
| `npm run build` | Type-check, build and validate bundle-size budgets |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run ESLint |
| `npm run format` | Format files using Prettier |
| `npm run format:check` | Verify formatting without changes |
| `npm test` | Run Vitest tests |
| `npm run test:watch` | Run Vitest in watch mode |
| `npm run test:e2e` | Run Playwright browser tests |

Typical verification sequence:

```bash
npm ci
npm run format:check
npm run lint
npm test
npm run build
npx playwright install chromium
npm run test:e2e
```

The build enforces compressed asset budgets of 180 KiB for JavaScript and 30 KiB for CSS. Browser tests cover authentication and key workspace/project/task journeys across desktop and mobile Chromium profiles, including accessibility checks.

## Deployment considerations

Build the production assets using `npm run build` and serve the generated static output from a web host with SPA route fallback. Configure the public API origin, HTTPS, cookie compatibility and CORS at the gateway. Do not embed secrets in Vite variables; frontend configuration is included in distributable browser assets.

## Related documentation

- [Platform overview](../README.md)
- [API gateway](../gateway-service/README.md)
- [Public OpenAPI contract](../gateway-service/api/openapi.json)
