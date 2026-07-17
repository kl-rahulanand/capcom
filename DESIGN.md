# Design System — Capcom

> "Mission control, quieted." Capcom = *Capsule Communicator*, the person in NASA
> mission control who talks to the astronauts. Calm, precise, unshowy authority.

## Product Context
- **What this is:** A runtime-agnostic AgentOps control plane. Operators import the
  durable state of each runtime instance, inspect agents/skills/effective access, watch
  freshness, and reconcile drift against desired state.
- **Who it's for:** SREs, platform engineers, and operators running fleets of AI agents.
- **Space/industry:** Dev-infra / observability / control planes. Peers by *feel*:
  Linear, Vercel, Railway, Temporal, PlanetScale. Peers by *domain*: Datadog, Grafana.
- **Project type:** Dark-first internal control-plane dashboard (embedded console served
  as static assets by the Go binary — vanilla HTML/CSS/JS, no build step, no framework).

## Aesthetic Direction
- **Direction:** Industrial / utilitarian, refined to Linear-grade polish.
- **Decoration level:** intentional (hairline borders, 1px top-highlight on elevated
  cards, faint accent glow on primary/active elements only, optional whisper of grid
  texture in empty/hero states). No gradients, no blobs, no decorative chrome.
- **Mood:** Fast, modern, effortless. Calm structured lists with instant drill-down —
  NOT a metrics cockpit / gauge wall. The tool's job is *inspect + control state*, so it
  reads like an issue tracker (Linear) more than a telemetry dashboard (Grafana).
- **Reference sites:** linear.app, vercel.com, railway.com, temporal.io.
- **First-3-seconds reaction:** "This is the sharpest, most modern ops tool I've used —
  and I can trust it with production agents."

## Typography
- **Display / Hero:** Geist 700 — modern grotesk built for product UI; tight tracking
  (`-0.02em`) at large sizes. Confident without shouting.
- **Body:** Geist 400 — same family for a tight, coherent system.
- **UI / Labels:** Geist 500/600.
- **Data / Tables:** Geist with `font-feature-settings: "tnum" 1` — tabular figures so
  metric and ID columns align.
- **HUD readout layer (signature):** **Geist Mono** — used prominently for eyebrows,
  metric values, agent IDs, endpoints, timestamps, and badges. This monospace layer is
  the mission-control "readout" and Capcom's distinguishing texture. Use it deliberately,
  not everywhere.
- **Code:** Geist Mono (alt: JetBrains Mono).
- **Loading:** Google Fonts for dev/preview; **self-host WOFF2 in the embedded console**
  so it works offline (the console is served by the Go binary — no external CDN at runtime).
- **Scale (px):** eyebrow 11 / small 12 / body 14 / lead 16 / h4 16 / h3 22 / h2 15(semibold)
  / h1 44 (marketing) · 22 (in-app view title). Line-height 1.5 body, 1.05–1.1 display.

## Color
- **Approach:** restrained — monochrome neutrals + **one signal accent**. Color is rare
  and meaningful. Red is reserved *strictly* for danger; never used as brand/primary.
- **Brand / signal accent:** `#3DE1A0` (dark) / `#0E9E6E` (light) — "live / nominal / go."
  Used for primary actions, active nav, focus rings, links, and the "active/live" status.
- **Neutrals (dark, canvas → detail):**
  - Canvas `#0A0C10` · Surface `#10131A` · Elevated `#161A22`
  - Hairline `#232833` · Soft line `#1A1F28`
  - Text `#E8ECF2` · Muted `#8A93A2` · Faint `#5B6473`
- **Neutrals (light):**
  - Canvas `#FBFBFD` · Surface `#FFFFFF` · Elevated `#FFFFFF`
  - Hairline `#E7E9EE` · Soft line `#EEF0F4`
  - Text `#0E1116` · Muted `#5B6472` · Faint `#98A0AD`
- **Semantic:**
  - success / nominal / active — `#3DE1A0` (light `#0E9E6E`)
  - warning / stale / cached — `#F2B441` (light `#B67C0B`)
  - **danger / failed / drift — `#F2555A` (light `#D23C42`)**
  - info — `#5AB0FF` (light `#2C7BE5`)
  - Each semantic color has a 10–12% alpha "-dim" fill for badge/alert backgrounds.
- **Dark mode:** dark is the default. Light theme is a full re-mapping (not a filter):
  paper surfaces, ink text, borders lightened, and the accent darkened to `#0E9E6E` so it
  passes contrast on white. Semantic colors darkened similarly for light.

## Spacing
- **Base unit:** 4px.
- **Density:** comfortable (looser than the current cramped 10px cells; tables and stat
  cards get real breathing room). Efficient, not cramped; effortless, not sparse.
- **Scale:** 2xs(2) xs(4) sm(8) md(12) lg(16) xl(24) 2xl(32) 3xl(48) 4xl(64).
- **Rhythm:** 8px baseline; card padding 18–22px; table cell padding 13px×16px.

## Layout
- **Approach:** grid-disciplined. App shell = left sidebar nav (Overview / Instances /
  Agents) + top bar (instance picker + refresh) + content. Detail opens in a right-side
  **drawer**, not a full-page navigation — fast drill-down, context preserved.
- **Grid:** stat cards 4-up (2-up ≤980px, 1-up ≤720px). Content max-width ~1180px on
  marketing/preview surfaces; the app itself is fluid within the shell.
- **Border radius (hierarchical):** controls/inputs/buttons `8px`, cards/panels/tables
  `12px`, pills/badges/status `999px`. Tables keep crisp internal cell edges.
  (Deliberate move away from the old razor-sharp 0–4px edges.)
- **Elevation:** `0 1px 0 rgba(255,255,255,.04) inset, 0 18px 40px rgba(0,0,0,.45)` (dark);
  soft `0 1px 2px + 0 12px 30px rgba(16,24,40,.06)` (light).

## Motion
- **Approach:** minimal-functional → intentional. Motion aids comprehension; never decorative.
- **Easing:** enter `ease-out`, exit `ease-in`, move `ease-in-out`.
- **Duration:** micro 100ms (hover/press), short 140ms (state changes, nav), medium 200ms
  (view switch fade+rise, drawer slide-in). No scroll-driven choreography.
- **Focus:** animated focus ring = 3px `signal-dim` halo + `signal-line` border.

## Anti-patterns (never ship)
- Red (or any semantic color) used as the brand/primary accent.
- Purple/violet gradients; gradient CTA buttons.
- 3-column icon-in-colored-circle feature grids; centered-everything layouts.
- Inter / Roboto / system-ui as the display or body font.
- Uniform bubble-radius on everything; a wall of gauges/charts on the Overview.

## Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-07-16 | Initial design system created | `/design-consultation`; landscape research on Linear/Vercel/Railway/Temporal. User chose "fast/modern/effortless" + "modern reinterpretation" + dark-first-with-light. |
| 2026-07-16 | Brand accent = aqua-green `#3DE1A0`; red reserved for danger | Old `#ef5350` brand red burned the one color operators need for failed/drift. Aqua-green is distinctive in the category and doubles as the "live/nominal" status. |
| 2026-07-16 | Geist + Geist Mono; monospace as a prominent HUD layer | Signature "readout" texture tied to the Capsule-Communicator concept; Geist gives modern grotesk polish with tabular figures for data. |
| 2026-07-16 | Calm drill-down lists over a metrics cockpit | Capcom's job is inspect + control *state*, closer to Linear/Vercel inventory than Grafana telemetry. |

## Artifacts
- Live preview (dogfoods the whole system, dark + light): saved under the gstack designs
  dir for this project — `design-system-20260716/capcom-design-preview.html` plus
  `preview-dark.png` / `preview-light.png`.
- Implementation target: `internal/api/ui/styles.css` (CSS-first refactor) and light
  markup adjustments in `internal/api/ui/index.html`. No JS logic changes required.
