# Taskflow web application

Phase 13 provides the browser foundation for the task-management platform. It is a strict TypeScript, React, and Vite application with a responsive shell and complete authentication journeys.

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

The production build fails if compressed JavaScript exceeds 180 KiB or CSS exceeds 30 KiB. Playwright exercises registration, login and route restoration, refresh-session restoration, logout, password reset, and automated accessibility on desktop and mobile Chromium profiles.

## Architecture

- `src/api` contains contract types, centralized error mapping, retry rules, cancellation support, and the cookie-aware API client.
- `src/features` owns authentication, shell, workspace, project, task, member, notification, and settings modules.
- `src/ui` contains accessible shared states, network feedback, branding, and the render error boundary.
- TanStack Query supplies the request cache. Server responses remain authoritative and failed mutations are never retried automatically.
- React Router guards authenticated routes and restores only same-origin relative return paths.

Access tokens exist only in memory. Refresh sessions use the gateway's `HttpOnly`, `SameSite` cookie, API calls include credentials, and logout/login events synchronize across tabs through `BroadcastChannel`. The UI never makes privileged authorization decisions.
