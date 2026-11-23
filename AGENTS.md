# Tablon — agent guide

Tablon is an open-source, lightweight, AI-first issue tracker: a Linear alternative for humans and coding agents. Go + SQLite backend, Svelte 5 frontend, REST + SSE, Go CLI.

## Hard rules

1. Do not commit, amend, push or open a PR without explicit approval.
2. Read only the docs the task needs. Do not load UI docs for backend work.
3. Keep changes small and reviewable; explain tradeoffs like a senior engineer.
4. Never invent design values. Tokens live in `frontend/src/app.css`; rules live in `docs/design-system.md`. Change both together.
5. No light theme in v0.1.
6. Accessibility is required: focus, labels, keyboard, reduced motion, mobile targets.
7. Do not add React, Radix, dnd-kit, TanStack, date-fns or similar to the Svelte app.
8. Measure performance claims; do not claim targets without a benchmark.

## Docs index

| Load when | File |
|---|---|
| UI, CSS, layout, color, motion, components, accessibility | `docs/design-system.md` |
| Design tokens and base styles | `frontend/src/app.css` |
| Frontend commands | `frontend/README.md` |
| Product scope, decisions and contracts | `docs/product/build-brief.md` |
| Architecture and data model | `docs/architecture.md` (planned) |

Add new docs to this index. Keep each doc short and focused; link instead of repeating.

## Repo map

- Go module: `github.com/pablogtrz/tablon`.
- `frontend/` — Svelte 5 + TypeScript + Vite + Tailwind v4.
- `cmd/server/`, `cmd/pm/`, `internal/` — Go server and CLI (planned).
- `migrations/` — SQLite migrations (planned).
- `docs/` — agent docs.
- `data/` — runtime database (gitignored).

## Commands

```bash
cd frontend && pnpm dev      # dev server, proxies /api to :8080
cd frontend && pnpm check    # types + lint
cd frontend && pnpm test     # unit tests
cd frontend && pnpm build    # static assets
go run ./cmd/server          # API on 127.0.0.1:8080
go test ./...                # backend + CLI tests
docker compose up --build    # container
```

## Working style

- Prefer platform APIs over dependencies: native `<dialog>`, container queries, `Intl`, CSS transitions.
- Keep the backend/frontend boundary clean: Go services, Svelte 5 + TypeScript, REST + SSE, SQLite.
- Verify UI at 360, 390, 768, 1024 and 1440px.
