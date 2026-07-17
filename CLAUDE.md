# Capcom

Runtime-agnostic AgentOps control plane (Go backend + embedded static console).
See [README.md](README.md) for architecture, run instructions, and API contract.

## Design System
Always read [DESIGN.md](DESIGN.md) before making any visual or UI decisions.
All font choices, colors, spacing, and aesthetic direction are defined there.
Do not deviate without explicit user approval.
In QA mode, flag any code that doesn't match DESIGN.md.

The **primary console is a Next.js + shadcn/ui app in `web/`** (App Router, TypeScript,
Tailwind v4, dark-first + light). It is a separate frontend that calls the Go JSON API
(bearer admin token; base URL via `NEXT_PUBLIC_CAPCOM_API_URL`). See
[web/README.md](web/README.md) to run it and
[docs/console-redesign/BUILD_SPEC.md](docs/console-redesign/BUILD_SPEC.md) for the spec.
Follow DESIGN.md tokens; use the existing `web/src/lib` hooks/types and `web/src/components/ui`
shadcn components — don't hand-roll UI primitives.

The old vanilla console in `internal/api/ui/` (`index.html`, `styles.css`, `app.js`,
`durable.js`, served embedded by the Go binary) is **legacy**, kept only so the Go build
keeps working; it is superseded by `web/` and can be retired later.
