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
| CLI commands, flags, output and retries | `cli/README.md` |
| Portable agent adapter and per-client install | `agent/Tablon.md`, `agent/README.md` |
| Project overview, stack and license | `README.md` |
| Product scope, decisions and contracts | `docs/product/build-brief.md` |
| API contract, endpoints and schemas | `docs/api/openapi.yaml` |
| Self-hosting, reverse proxy, backup, upgrade | `docs/self-hosting.md` |
| Performance method and measured results | `docs/performance.md` |
| Tagging, artifacts and release runbook | `docs/release.md` |
| Architecture and data model | `docs/architecture.md` (planned) |

Add new docs to this index. Keep each doc short and focused; link instead of repeating.

## Repo map

- Go module: `github.com/pablogtrz/tablon`.
- `assets.go` — `go:embed` for `frontend/dist` and the OpenAPI spec.
- `frontend/` — Svelte 5 + TypeScript + Vite + Tailwind v4.
- `cmd/server/`, `internal/` — Go server.
- `cli/` — `tablon` CLI (`cli/tablon` entrypoint, `cli/README.md`).
- `migrations/` — SQLite migrations.
- `docs/` — agent docs; `docs/api/` holds the OpenAPI contract.
- `scripts/` — benchmark and smoke-test helpers.
- `.github/workflows/` — CI and release automation.
- `Dockerfile`, `docker-compose.yml`, `.env.example` — self-hosting.
- `agent/` — portable `Tablon.md` adapter and client wrappers.
- `data/` — runtime database (gitignored).

## Commands

```bash
cd frontend && pnpm dev      # dev server, proxies /api to :8448
cd frontend && pnpm check    # types + Svelte diagnostics
cd frontend && pnpm test     # unit tests
cd frontend && pnpm test:e2e # Chromium, Firefox and WebKit
cd frontend && pnpm build    # static assets for go:embed
go run ./cmd/server          # API on 127.0.0.1:8448; pair with Vite
go run ./cli/tablon --help   # CLI
go test ./...                # backend + CLI tests
go vet ./...
go test -race ./...
./scripts/bench.sh           # macOS startup/RSS/latency measurements
./scripts/bench-runtime.sh   # native vs Docker runtime/storage (Linux)
python3 scripts/bench-tokens.py --url <server> --cli <tablon>  # token cost
docker compose up --build    # container
```

A release/Docker binary needs `pnpm --dir frontend build` and `-tags embed`:

```bash
pnpm --dir frontend build
go build -tags embed -o bin/tablond ./cmd/server
```

`go test ./...` uses the empty `!embed` stub and works on a clean clone.

## Working style

- Prefer platform APIs over dependencies: native `<dialog>`, container queries, `Intl`, CSS transitions.
- Keep the backend/frontend boundary clean: Go services, Svelte 5 + TypeScript, REST + SSE, SQLite.
- Verify UI at 360, 390, 768, 1024 and 1440px.
