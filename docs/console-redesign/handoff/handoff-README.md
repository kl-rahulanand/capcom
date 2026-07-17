# Handoff: Capcom Console

## Overview
Capcom is a control plane for fleets of AI agents running on heterogeneous runtimes. It imports durable state from runtime **adapters** (Gantry, LangGraph, Temporal, Letta, CrewAI), each of which connects one or more **instances** (deployed environments), each running **agents** with skills and resolved access grants. The console gives operators: a fleet Overview (per-adapter health cards + "needs attention" queue), an adapter detail view (instances as collapsible groups with their agents), a fleet-wide Agents table, per-instance/per-adapter re-import actions, a ⌘K command palette, and a dark/light theme toggle.

## About the Design Files
The file in this bundle (`Capcom App.dc.html`) is a **design reference created in HTML** — an interactive prototype showing intended look and behavior, NOT production code to copy directly. Your task is to **recreate this design in the target codebase's existing environment** (React, Vue, etc.) using its established patterns, router, and component libraries — or, if no frontend exists yet, pick the most appropriate framework and implement the design there. The prototype's inline styles and mock data are a spec, not an implementation.

Open the .dc.html file in a browser to click through the prototype. All markup lives between `<x-dc>` tags with inline styles; the logic class at the bottom (`class Component`) holds the mock data model and all state transitions — read it as the behavioral spec.

## Fidelity
**High-fidelity.** Colors, typography, spacing, radii, and interactions are final. Recreate pixel-perfectly, mapping values onto your design-token system.

## Information Architecture
```
Sidebar
├── Overview            (default route)
├── Adapters ▾          (collapsible section header — collapsed by default,
│   ├── Gantry           auto-expands when an adapter route is active)
│   ├── LangGraph        each row: status dot + name + instance count
│   ├── Temporal
│   ├── Letta
│   └── CrewAI
└── Agents              (fleet-wide agent table)
```

## Design Tokens
Two themes, implemented as CSS custom properties on the app root. Dark is default.

| Token | Dark | Light | Used for |
|---|---|---|---|
| --cv | #0A0C10 | #FBFBFD | canvas / page background |
| --sf | #10131A | #FFFFFF | surface (sidebar, tables, cards' container) |
| --el | #161A22 | #FFFFFF | elevated (cards, palette) |
| --hl | #232833 | #E7E9EE | hairline border (strong) |
| --sl | #1A1F28 | #EEF0F4 | subtle border / hover fill / row separator |
| --tx | #E8ECF2 | #0E1116 | primary text |
| --mu | #8A93A2 | #5B6472 | muted text |
| --fa | #5B6473 | #98A0AD | faint text (labels, hints) |
| --ac | #3DE1A0 | #0E9E6E | accent (healthy, primary actions, links) |
| --acd | rgba(61,225,160,.11) | rgba(14,158,110,.10) | accent dim (badge bg, active nav bg) |
| --wn | #F2B441 | #B67C0B | warning (stale) |
| --wnd | rgba(242,180,65,.12) | rgba(182,124,11,.12) | warning dim |
| --dg | #F2555A | #D23C42 | danger (failed) |
| --dgd | rgba(242,85,90,.12) | rgba(210,60,66,.10) | danger dim |
| --onac | #06231A | #FFFFFF | text on accent buttons |
| --glow | rgba(61,225,160,.2) | rgba(14,158,110,.18) | primary-button glow shadow |
| --shdw | 0 1px 0 rgba(255,255,255,.04) inset, 0 18px 40px rgba(0,0,0,.45) | 0 1px 2px rgba(16,24,40,.06), 0 12px 30px rgba(16,24,40,.10) | floating surfaces (palette) |
| --chi | 0 1px 0 rgba(255,255,255,.04) inset | 0 1px 2px rgba(16,24,40,.05) | card inner highlight |

**Typography.** UI font: **Geist** (Google Fonts, 400/500/600/700). Data font: **Geist Mono** (400/500/600) — used for ALL identifiers (instance names, agent names, access chips), numbers/metrics, timestamps, eyebrow labels, keyboard hints, and the CAPCOM wordmark. Numeric spans use `font-feature-settings: 'tnum' 1`.

Scale: page title 22px/700/-0.02em · card title 15px/600 · big metric 26px/600 mono · body/rows 13px · row identifiers 13px/500 mono · secondary 12px · eyebrow/column headers 11px mono, uppercase, letter-spacing .08em, color --fa · badges/chips 11px mono · line-height 1.5.

**Spacing & shape.** Radii: 12px cards/tables, 8px buttons/nav items, 6px kbd chips, 999px pills/dots. Content padding: 26px 32px. Sidebar: 230px fixed. Topbar: 56px fixed, 1px --sl bottom border. Table row padding: 12px 18px; nested agent rows indent to 44px left.

**Status language** (used consistently everywhere): ok/healthy → --ac dot, "healthy" badge · stale → --wn dot, "N stale" badge · failed → --dg dot, "import failed" badge. Status dots: 8px circles (6px in sidebar/agent lists); "live" dots get a 3px halo of the dim color (`box-shadow: 0 0 0 3px <dim>`).

## Screens

### 1. Sidebar (persistent)
- Header: 8px mint dot with --acd halo + "CAPCOM" (mono 14px/600, letter-spacing .14em); "control plane" beneath (mono 11px --fa).
- Nav items: 13px/500, padding 7px 10px, radius 8px. Active = --acd bg + --ac text. Hover = --tx text.
- "Adapters" section header: mono 11px uppercase eyebrow with ▸/▾ chevron and total count; click toggles the adapter list. Collapsed by default; forced open while an adapter view is active.
- Adapter rows (when open): 6px status dot + name (13px) + instance count (mono 11px, right-aligned). Active row: --sl bg, --tx text.
- "Agents" item with fleet agent count.
- Footer (pinned bottom): system status line — mono 11px, dot + text; color/text derived from worst status: "all systems normal" (--ac) or "N items need attention" (--wn/--dg). Below: "v0.4.2 · go/capcom" (mono 11px --fa).

### 2. Topbar (persistent, 56px)
- Left: on Overview/Agents an environment picker chip ("env All environments ▾", 1px --hl border, radius 8px); on adapter detail a breadcrumb "Adapters › {Name}" (parent --fa + clickable → Overview, current --tx/500).
- Right: "last update {rel}" (mono 12px --fa) · "↻ Refresh now" outline button (hover: --ac border+text) · "⌘K" kbd chip (opens palette) · theme toggle "◐ dark"/"◑ light" (mono 12px outline chip).

### 3. Overview
- Header row: title "Overview" + subtitle "Everything Capcom is watching, grouped by runtime adapter. Click a card to drill in." + primary button **"+ Add instance"** (accent bg, --onac text, 8px radius, glow shadow, hover brightness 1.08).
- **Adapter cards**: 3-column grid (`repeat(3, minmax(0,1fr))`, 16px gap). Card: --el bg, 1px --hl border, 12px radius, 20px padding, --chi shadow; hover: border-color → --ac (.14s ease-out); whole card navigates to adapter detail.
  - Row 1: status dot · adapter name (15px/600) · status pill right.
  - Row 2: two metrics side-by-side (32px gap): instance count over "instances connected", agent count over "agents running" (26px mono over 12px --mu).
  - Footer (12px, above 1px --sl separator): healthy → "All state fresh · updated {rel}" in --mu; degraded → the specific problem in the status color (e.g. "edge-sydney not updated for 14m"); right: "View details →" 12px --fa.
  - Last grid cell: **"Connect an adapter"** ghost card — 1px dashed --hl, centered "+" + label + "Point Capcom at another runtime", hover → --ac.
- **"Needs your attention" table**: surface container. Header: title + count ("2 items" / "all clear"). Row: halo dot · `{instance} · {Adapter}` (mono 13px) · plain-language message (12px --mu, flex:1) · status pill · action link in --ac ("Retry import" / "Re-import"). Messages: failed → "Couldn't import state — connection refused. Check the endpoint."; stale → "State is {age} old — older than the 5m freshness budget." Empty state: single row, mint halo dot + "Nothing needs attention right now — all instances are fresh."

### 4. Adapter detail (e.g. Gantry)
- Header: name (22px/700) + status pill · subtitle "4 instances connected · 59 agents running · state re-imported automatically every 30s" · buttons right: **"+ Add instance"** (outline, hover --ac), "Adapter settings" (outline), **"↻ Re-import all instances"** (primary accent).
- **Instance groups** — one collapsible card per instance, 14px gap:
  - Header row (14px 18px padding, hover --sl, click toggles): ▸/▾ chevron (11px --fa) · 8px status dot · instance name (mono 14px/600) · env pill ("production"/"staging"/"development", mono 11px, 1px --hl border) · "{N} agents · {M} skills" (12px --mu) · right: "updated {rel}" (12px; --mu when ok, status color otherwise; failed reads "import failed · 62s ago") · "↻ Re-import" outline button (--ac text, hover --ac border). Re-import clicks must not toggle the group (stopPropagation).
  - Expanded body: 4-column agent table, columns `2fr .7fr 2.3fr 1fr` = Agent / Skills / Can access / Status. Column headers: 11px mono uppercase eyebrow. Rows: agent name (mono 13px/500) · skill count (mono 13px) · access chips (mono 11px pills, --sl bg + --hl border, wrap with 6px gap; empty → "none resolved" in --fa) · status pill ("running" accent / "idle" faint), left-aligned in cell.
  - If not all agents shown: footer note "Showing {n} of {N} agents" (12px --fa).
  - First instance expanded by default; others collapsed.
- Page footer note (12px --fa): "Connection: grpc://gantry.agents.internal:7233 · adapter v1.2.0 · freshness budget 5m per instance".

### 5. Agents (fleet-wide)
- Title "Agents" + subtitle "Every agent Capcom has imported, across all adapters and instances."
- Single table, columns `1.6fr 1.8fr .6fr 2fr .9fr` = Agent / Adapter · Instance / Skills / Can access / Status. Location cell: "Gantry · gantry-prod-us" (mono 12px --mu). Same chip/pill treatments as above.
- Footer note: "Showing {n} of {N} agents — the full list loads from each runtime snapshot."

### 6. Command palette (⌘K)
- Scrim: rgba(4,6,9,.55) + backdrop-blur 1.5px, click closes.
- Panel: 600px wide, centered, top 110px; --el bg, 1px --hl border, 12px radius, --shdw; enters with rise animation.
- Input row: "›" prompt + placeholder "Type a command…" with a 2px mint caret block + "esc" kbd chip (click closes).
- "Actions" eyebrow, then rows (radius 8px, hover --sl): "↻ Re-import all stale or failed instances *N pending*" (runs fix + closes) · "→ Open Gantry adapter" · "↻ Refresh all adapters now". Footer strip: "↵ run · esc close · capcom ⌘K" (mono 11px --fa).
- Keyboard: ⌘K/Ctrl+K toggles (preventDefault), Escape closes.

## Interactions & Behavior
- **Navigation**: sidebar items, adapter cards, and breadcrumb all switch views. View switch plays `rise` animation: fade + translateY(4px→0), .2s ease-out.
- **Re-import** (per-instance, per-adapter "all", attention-row action, palette bulk action): sets the instance status → ok and "updated just now"; attention list and all badges/dots/footers recompute immediately; topbar "last update" → "just now".
- **Refresh now**: updates "last update" only.
- **Theme toggle**: swaps the full token set (table above); label flips ◐ dark / ◑ light. Persist choice.
- **Hovers**: rows/cards → --sl fill or --ac border; buttons → brightness(1.08) (primary) or --ac border/text (outline). Transitions ~.15–.2s ease-out.
- All status displays derive from one source of truth (instance.status + instance.updated) — never store badge text separately.

## State Management
- `view`: 'overview' | 'agents' | adapter id — router in a real app (`/`, `/agents`, `/adapters/:id`).
- `adaptersOpen`: sidebar section toggle (default: open iff current route is an adapter).
- `expanded`: map of "{adapterId}:{instanceName}" → bool (default: first instance of each adapter true).
- `palette`: bool, global ⌘K listener.
- `theme`: 'dark' | 'light'.
- `lastUpdate`: relative timestamp string (real app: derive from fetch time, tick every ~10s).
- Data shape: `Adapter { id, name, endpoint, ver, poll, instances: Instance[] }`; `Instance { name, env, skills, total, updated, status: 'ok'|'stale'|'failed', agents: Agent[] }`; `Agent { name, skills, access: string[] ("payments:rw"), status: 'running'|'idle' }`. Mock fleet data is in the logic class — reuse it as fixture data.

## Assets
No images or icon files. Glyphs are text: ↻ ▸ ▾ › → ↵ ◐ ◑ ✕ +. Fonts from Google Fonts: Geist, Geist Mono.

## Files
- `Capcom App.dc.html` — the full interactive prototype (markup between `<x-dc>` tags; behavior + mock data in the `Component` logic class at the end of the file).
