# Capcom Console — Next.js + shadcn Rebuild: Build Spec

**Audience:** the implementing agent (Codex). This is the authoritative brief. Where this
doc and a referenced file disagree on API/behavior, the referenced source file wins and you
flag it.

## 0. Goal in one line
Replace the vanilla embedded console with a **separate Next.js + shadcn/ui frontend** that
faithfully implements the approved design handoff, wired to the **existing Go JSON API**,
covering **every feature the current console has** — nothing regresses.

## 1. Hard constraints
- **Stack:** Next.js (App Router) + TypeScript (strict) + Tailwind CSS + shadcn/ui +
  lucide-react + next-themes. Data layer: TanStack Query. Fonts via `geist` package
  (`geist/font/sans`, `geist/font/mono`) — do NOT use Inter/system fonts.
- **Separation:** the Next.js app is its own project at repo path `web/`. The Go binary stays
  the backend API. They are separate origins in dev (`:3000` frontend, `:8080` API).
- **Do not break the Go build or tests.** Leave the existing embedded console
  (`internal/api/ui/*`) in place and working. The only Go change allowed in this effort is
  adding CORS (see §7). Run `go build ./...` and `go test ./...` before you finish.
- **No secrets in the repo.** Admin token is entered at runtime by the user (see §6).
- **Design fidelity is high.** Colors, type, spacing, radii, and interactions are final —
  match the handoff pixel-for-pixel, mapped onto shadcn tokens.

## 2. Read these first (sources of truth)
- Design system: `DESIGN.md` (repo root).
- Visual + behavioral spec: `docs/console-redesign/handoff/handoff-README.md` and the
  interactive prototype `docs/console-redesign/handoff/capcom-app.dc.html`. The `<x-dc>`
  markup is the layout; the `class Component` block at the bottom is the behavioral spec
  (state, transitions, derived status). Treat its mock data shape as the fixture shape.
- **API contract:** `api/openapi.yaml` and the Go handlers in `internal/api/router.go`
  (routes + JSON response structs are defined there). These win over anything else for
  endpoint paths and field names.
- Existing console behavior to preserve: `internal/api/ui/index.html`, `app.js`, `durable.js`.

## 3. Design tokens → Tailwind + shadcn
Implement BOTH themes as CSS custom properties on the root, dark = default, via `next-themes`
(`class` strategy, persisted). Use the exact values from `DESIGN.md` / the handoff token table.
Wire shadcn's semantic variables to our palette so components inherit it:
- `--background` = canvas `#0A0C10` / `#FBFBFD`; `--card`/`--popover` = elevated
  `#161A22` / `#FFFFFF`; surfaces (sidebar/tables) = `#10131A` / `#FFFFFF`.
- `--border` = `#232833` / `#E7E9EE`; subtle border/hover fill = `#1A1F28` / `#EEF0F4`.
- `--foreground` = `#E8ECF2` / `#0E1116`; `--muted-foreground` = `#8A93A2` / `#5B6472`;
  faint labels `#5B6473` / `#98A0AD`.
- `--primary` = accent `#3DE1A0` / `#0E9E6E`; `--primary-foreground` = `#06231A` / `#FFFFFF`.
- Status: healthy/ok = accent; warning/stale = `#F2B441` / `#B67C0B`;
  danger/failed = `#F2555A` / `#D23C42`; each with the dim (~.10–.12 alpha) badge bg.
- Radii: cards/tables 12px, buttons/nav 8px, kbd chips 6px, pills/dots 999px.
- Type: Geist for UI; **Geist Mono for ALL identifiers, numbers/metrics, timestamps,
  eyebrow/column labels, kbd hints, and the CAPCOM wordmark**; numeric spans use tabular
  figures (`font-feature-settings:'tnum' 1`). Scale per handoff (§Typography).
- Glyphs are text (↻ ▸ ▾ › → ↵ ◐ ◑ ✕ +) or lucide equivalents — no icon image assets.

## 4. Information architecture / routes (App Router)
- `/` — **Overview**: per-adapter health cards (3-col grid) + "Connect an adapter" ghost card
  + "Needs your attention" queue. Card click → adapter detail.
- `/adapters/[adapterId]` — **Adapter detail**: header (name, status pill, subtitle, actions),
  instances as collapsible groups each containing an agent sub-table; per-instance and
  "re-import all" actions; page footer note (endpoint/version/freshness).
- `/agents` — **Agents (fleet-wide)** table across all adapters/instances.
- Persistent **sidebar** (CAPCOM wordmark + "control plane"; Overview; collapsible **Adapters**
  section listing each adapter with status dot + instance count, auto-expanded on an adapter
  route; Agents with fleet count; footer system-status line + `v… · go/capcom`).
- Persistent **topbar** (56px): env picker chip on Overview/Agents / breadcrumb on detail;
  right: "last update {rel}", "↻ Refresh now", "⌘K" chip, theme toggle.
- **⌘K command palette** (shadcn `command` / cmdk in a dialog): actions from the prototype
  (re-import stale/failed, open an adapter, refresh all). ⌘K/Ctrl+K toggles, Esc closes.

## 5. "Adapter" mapping (there is NO adapter entity in the API)
An **adapter = a distinct `runtime_type`**. Build the adapters model client-side:
- Fetch `GET /v1/runtime-instances` (list of runtime connections; each has `id`, `name`,
  `display_name`, `environment`, `runtime_type`, `mode`, `status`, `endpoint`,
  `last_synced_at`, `last_sync_status`, `sync_interval_seconds`, `last_error`, …).
- Group by `runtime_type` → adapters (Gantry, LangGraph, Temporal, Letta, CrewAI, …). Each
  group's `instances` are its runtime-instances. `adapterId` in the route = the runtime_type.
- Per instance, agent count / skills come from persisted state (see §6 endpoints).
- Derive adapter/instance status + "needs attention" from `status` + freshness
  (`last_synced_at` vs a freshness budget; mirror the handoff's ok/stale/failed language).
- Adapters that have no live connection yet simply don't appear; the "Connect an adapter"
  card starts the add-instance flow.

## 6. API integration
- **Base URL:** `process.env.NEXT_PUBLIC_CAPCOM_API_URL` (default `http://127.0.0.1:8080`).
- **Auth:** every `/v1/*` request needs `Authorization: Bearer <adminToken>`. Mirror the
  current console: a **Connection dialog** captures the admin token, stored in
  `sessionStorage` (cleared on tab close), used for all requests. On `401`, clear it and
  reopen the dialog with a message. `GET /healthz` needs no token (used for sidebar status).
- Use a typed API client + TanStack Query hooks. Generate TS types from the response structs
  in `internal/api/router.go` / `api/openapi.yaml`.
- **Endpoints you will use** (confirm exact paths/fields in `router.go`):
  - `GET /healthz` → `{status, service, version}`.
  - `GET /v1/runtime-instances` / `GET /v1/runtime-instances/{id}` — instances.
  - `POST /v1/runtime-instances` — create instance (add-instance flow; needs name,
    runtime_type, mode, endpoint, auth_ref, actor, reason, …).
  - `POST /v1/runtime-instances/{id}/test` → `{status, message, capabilities, metadata}`.
  - `POST /v1/runtime-instances/{id}/sync` (body `{actor, reason}`) — the "Re-import" action.
  - `GET /v1/runtime-instances/{id}/sync-runs` — sync history.
  - `GET /v1/runtime-instances/{id}/agents` — persisted agents for an instance.
  - `GET /v1/runtime-instances/{id}/subagent-executions` — delegated subagent executions.
  - `GET /v1/agents` / `GET /v1/agents/{id}` — persisted agents (fleet + detail).
  - `GET /v1/agents/{id}/skills` — agent skills (name, description, source, status, version,
    tool_ids, workflow_refs).
  - `GET /v1/agents/{id}/access` — effective access (`selections[]`: kind, id, name, allowed,
    attributes).
  - `POST /v1/agents/{id}/actions/reconcile-access` (body `{selections, actor, reason,
    idempotency_key, dry_run}`) — the reconcile control action (control_enabled only).

## 7. The one backend change: CORS (Go)
Because the frontend is now a separate origin, add CORS to `internal/api/router.go`
(wrap the handler alongside `requestLogger`/`adminAuth`):
- Allowed origins from a new config env (e.g. `CAPCOM_CORS_ALLOWED_ORIGINS`, comma-separated;
  default `http://localhost:3000,http://127.0.0.1:3000`). Read via `internal/config`.
- Handle `OPTIONS` preflight (204) for `/v1/*`.
- Allow headers `Authorization, Content-Type`; methods `GET, POST, PATCH, OPTIONS`.
- Do not weaken auth. Add/extend a unit test in `internal/api/router_test.go` for preflight +
  an allowed-origin response header. Keep all existing tests green.
- Document the new env var in `README.md` config table.

## 8. Feature-coverage matrix — NOTHING regresses
Every row must be present and working against the real API:

| Feature (today) | Where it lives in the rebuild |
|---|---|
| Admin-token auth + 401 handling | Connection dialog (shadcn Dialog), sessionStorage |
| Runtime instances list/select | Sidebar adapters→instances + Overview/adapter views |
| Test connection → capabilities | Adapter detail (per instance): capability chips + status |
| Sync requiring actor + reason | "Re-import" opens a dialog (prefill actor=`local-operator`, editable reason); on submit calls `/sync` |
| Persisted agents table (name, kind, runtime_agent_id, status, freshness, description) | Adapter-detail instance sub-tables + `/agents` fleet table |
| Agent search | Search input on `/agents` |
| Agent drill-down: overview, skills (desc/tools/workflows/version/source), effective access | **Agent drawer** (shadcn Sheet) opened from any agent row |
| Reconcile-access editor (dry-run/apply, idempotency_key) | Inside the agent drawer, shown only when instance `mode == control_enabled` |
| Subagent (delegated) executions | A section/tab on adapter detail (and/or instance group): subagent_type, owner, status, parent_run, task, observed |
| Health / API status | Sidebar footer system-status line |
| Freshness semantics (live/cached/stale) + status (active/failed) | Status dots + badges everywhere, one source of truth = instance.status + last_synced_at |

## 9. Interactions (from the handoff)
- View switches play the `rise` animation (fade + translateY 4px→0, .2s ease-out).
- Re-import (per-instance / per-adapter / attention-row / palette bulk): calls sync, then
  refetches; badges/dots/footers/attention recompute from fresh data; topbar "last update"
  updates. Re-import inside a collapsible header must `stopPropagation` (not toggle the group).
- Refresh now: refetch + update "last update".
- Theme toggle: swap token set; persist; label ◐ dark / ◑ light.
- Hovers/transitions ~.14–.2s ease-out; primary buttons brightness(1.08); outline → accent border.
- Loading states: shadcn Skeleton for cards/tables while queries are pending.
- Toasts (shadcn Sonner) for sync/reconcile results and errors (mirror current `showNotice`).

## 10. Dev workflow + docs to produce
- `web/README.md`: how to run (`npm install`, `npm run dev` on :3000, point at API via
  `NEXT_PUBLIC_CAPCOM_API_URL`), how to build, env vars.
- Update root `README.md`: note the new `web/` frontend, the CORS env var, and that the
  console is now a separate Next.js app (keep the existing embedded-console note until it's
  formally retired — do not delete it in this pass).
- Add `web/` build artifacts to `.gitignore` (`node_modules`, `.next`, etc.).

## 11. Definition of done
- `web/` app runs, all three routes + sidebar/topbar/palette/theme work.
- Every row in §8 works end-to-end against a running Go API (with a token).
- Design matches the handoff (spot-check dark + light).
- `go build ./...` and `go test ./...` pass; CORS test added.
- TypeScript strict passes (`tsc --noEmit`) and `npm run build` (or `next build`) succeeds.
- READMEs updated. No secrets committed.
