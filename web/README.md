# Capcom Console (web)

The Capcom operator console — a Next.js (App Router) + shadcn/ui frontend that talks to
the Go API. Design system lives in the repo-root [DESIGN.md](../DESIGN.md); build spec in
[docs/console-redesign/BUILD_SPEC.md](../docs/console-redesign/BUILD_SPEC.md).

## Stack
- Next.js 16 (App Router) + React 19, TypeScript strict
- Tailwind v4 (CSS-first) + shadcn/ui (Base UI) + lucide-react
- next-themes (dark default + light), TanStack Query, Geist / Geist Mono

## Run locally
The frontend is a separate app from the Go backend (different origins in dev).

1. Start the Go API (from the repo root): `make run` — listens on `:8080`.
   Ensure `CAPCOM_CORS_ALLOWED_ORIGINS` includes this app's origin (the default already
   allows `http://localhost:3000`).
2. Start the console:

   ```bash
   cd web
   npm install      # first time only
   npm run dev      # http://localhost:3000
   ```

3. Open http://localhost:3000, enter your `CAPCOM_ADMIN_TOKEN` in the Connection dialog
   (kept in sessionStorage, cleared when the tab closes).

## Configuration
| Variable | Default | Description |
|---|---|---|
| `NEXT_PUBLIC_CAPCOM_API_URL` | `http://127.0.0.1:8080` | Base URL of the Capcom Go API |

Set it in `web/.env.local` if your API runs elsewhere.

## Build
```bash
npm run build      # production build
npx tsc --noEmit   # type check
```

## Structure
- `src/app/` — routes: `/` (Overview), `/adapters/[adapterId]` (adapter detail),
  `/agents` (fleet agents)
- `src/components/` — app shell, overview, adapter detail, agents fleet, agent drawer,
  dialogs, and `ui/` (shadcn components)
- `src/lib/` — typed API client, auth, TanStack Query hooks, and the client-side
  adapters model (groups runtime instances by `runtime_type`)
