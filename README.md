<p align="center">
  <img src="frontend/public/favicon.svg" width="112" height="112" alt="Tablon logo" />
</p>

<h1 align="center">Tablon</h1>

<p align="center">
  <strong>Lightweight, AI-first issue tracking for humans and coding agents.</strong>
</p>

<p align="center">
  <a href="LICENSE"><img alt="License: AGPL-3.0-only" src="https://img.shields.io/badge/license-AGPL--3.0-FFA443?style=flat-square&labelColor=0A0A0B"></a>
  <a href="https://go.dev"><img alt="Go 1.27+" src="https://img.shields.io/badge/go-1.27%2B-8AB4FF?style=flat-square&labelColor=0A0A0B&logo=go&logoColor=white"></a>
  <a href="https://svelte.dev"><img alt="Svelte 5" src="https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&labelColor=0A0A0B&logo=svelte&logoColor=white"></a>
  <a href="https://sqlite.org"><img alt="SQLite WAL" src="https://img.shields.io/badge/SQLite-WAL-003B57?style=flat-square&labelColor=0A0A0B&logo=sqlite&logoColor=white"></a>
  <a href="agent-instructions/AGENTS.md"><img alt="AI-first: agents" src="https://img.shields.io/badge/AI--first-agents-67D4F5?style=flat-square&labelColor=0A0A0B"></a>
  <a href="docs/product/build-brief.md"><img alt="Self-hosted" src="https://img.shields.io/badge/self--hosted-yes-6EE7B0?style=flat-square&labelColor=0A0A0B"></a>
  <a href="docs/product/build-brief.md"><img alt="Status: v0.1 in development" src="https://img.shields.io/badge/status-v0.1%20in%20development-F5C451?style=flat-square&labelColor=0A0A0B"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#features">Features</a> ·
  <a href="#agent-workflow">Agent workflow</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#development">Development</a> ·
  <a href="#license">License</a>
</p>

---

**Tablon** is a self-hosted issue tracker for small teams and coding agents. It keeps the workflow clear and the footprint small: a Go + SQLite backend, a calm dark-only Svelte UI, REST + SSE, and a `tablon` CLI built for token-efficient automation.

> Linear-like clarity. Self-hosted simplicity. Agent-ready by default.

## Why Tablon?

- **Small by design.** One Go service, SQLite in WAL mode, and no accounts, organizations or role system.
- **Agent native.** Agents discover, claim, update and comment through a compact CLI or the REST API.
- **Live without ceremony.** Per-project SSE keeps boards, issues and comments in sync; reconnect and refetch when needed.
- **Readable work.** Markdown descriptions and comments with a CodeMirror editor, headings, code blocks and copyable previews.
- **Fast to scan.** Dense issue tables, a four-column board, status at a glance and no decorative noise.
- **Dark only.** A deliberate dark interface with amber accents, accessible focus and reduced-motion support.

## Features

| Area | What you get |
|---|---|
| Projects and epics | Project identifiers, statuses, gradient/emoji logos and epic tracking |
| Issues | `PROJ-123` identifiers, statuses, priorities, epics, agent assignment and filters |
| Board | Todo, In Progress, In Review and Done columns with drag, keyboard and menu moves |
| Comments | Chronological Markdown comments with edit/delete and agent attribution |
| Editor | CodeMirror 6 with live Markdown styling, fenced code blocks and copy actions |
| Automation | `tablon` CLI, JSON/compact output, idempotent creates, claim/release |
| Live updates | Per-project SSE with heartbeat, reconnect/refetch and polite announcements |
| Security | Optional bearer-token auth; `/health` and `/ready` stay public |
| Deployment | One binary with embedded SPA, Docker Compose, non-root read-only container |
| API contract | OpenAPI 3.0.3 served at `/openapi.yaml` |

## Quick start

### Docker Compose (recommended)

Prerequisites: Docker with Compose v2.

```bash
cp .env.example .env
# Set TABLON_API_TOKEN before exposing Tablon beyond 127.0.0.1.
docker compose up -d --build
```

Open **http://127.0.0.1:8080**. Compose publishes only on loopback by default and
stores SQLite in the `tablon-data` named volume.

```bash
docker compose logs -f
docker compose down          # keeps the volume
```

### Prebuilt binary

Download `tablond` for your OS/arch from
[GitHub Releases](https://github.com/pablogtrz/tablon/releases/latest), verify
`SHA256SUMS`, then run it:

```bash
TABLON_DATA_DIR=./data ./tablond
```

Open **http://127.0.0.1:8080**. Install the CLI from the same release or with
`go install github.com/pablogtrz/tablon/cmd/tablon@latest`.

### From source (development)

Prerequisites: Go 1.27+, Node 22.12+ and pnpm 11+.

```bash
git clone https://github.com/pablogtrz/tablon.git
cd tablon
go run ./cmd/server
```

In a second terminal:

```bash
cd frontend
pnpm install
pnpm dev
```

Open **http://127.0.0.1:5173**. Vite serves the UI and proxies `/api` to the Go
server on `:8080`; `go run` without `-tags embed` serves the API only.

The server embeds the built SPA when compiled with `-tags embed`, so production
and Docker need no Node runtime. The Vite dev server is only for frontend
development:

```bash
pnpm --dir frontend build
go build -tags embed -o bin/tablond ./cmd/server
```

See [`docs/self-hosting.md`](docs/self-hosting.md) for reverse proxies, token
guidance, backups and upgrades.

## Agent workflow

```bash
# Install the CLI
go install github.com/pablogtrz/tablon/cmd/tablon@latest

# Point it at the server
export TABLON_CLI_URL="http://127.0.0.1:8080"
export TABLON_CLI_TOKEN="optional-token"

# Discover work
tablon project list --compact
tablon issue list --project TABLON --status todo --compact

# Claim, work, update, comment, release
tablon issue claim TABLON-17 --agent my-agent
tablon issue get TABLON-17
tablon issue update TABLON-17 --status in-progress
tablon issue comment TABLON-17 --agent my-agent --body "Implementation started"
tablon issue update TABLON-17 --status in-review
tablon issue release TABLON-17 --agent my-agent
```

Use `--compact` for routine reads and `--json` when parsing. The ready-to-copy template for target repositories lives in [`agent-instructions/AGENTS.md`](agent-instructions/AGENTS.md).

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `TABLON_HOST` | `127.0.0.1` | API bind address |
| `TABLON_PORT` | `8080` | API port |
| `TABLON_DATA_DIR` | `./data` | SQLite database directory |
| `TABLON_API_TOKEN` | empty | Optional bearer token for `/api/v1` |
| `TABLON_IDLE_ENABLED` | `true` | Enable the idle activity gate |
| `TABLON_IDLE_TIMEOUT` | `5m` | Quiet period before idle mode |
| `TABLON_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `TABLON_CLI_URL` | `http://127.0.0.1:8080` | CLI API base URL |
| `TABLON_CLI_TOKEN` | empty | CLI bearer token |
| `TABLON_HOST_PORT` | `8080` | Compose-only host port published on `127.0.0.1` |

Native binds default to loopback. If you expose Tablon beyond your machine, set
`TABLON_API_TOKEN` and use HTTPS through a reverse proxy. The API contract is
served at `/openapi.yaml`.

## Architecture

| Layer | Choice |
|---|---|
| Backend | Go 1.27, `net/http`, pure-Go SQLite, embedded migrations and SPA |
| Frontend | Svelte 5, TypeScript, Vite, Tailwind CSS v4, CodeMirror 6 |
| Contract | REST + SSE, OpenAPI 3.0.3, structured JSON errors, cursor pagination |
| CLI | Go, REST only, no direct database access |
| Deployment | One binary or Docker Compose; no Node at runtime |
| Storage | SQLite WAL, per-project issue counters, cascade deletes |

More detail lives in [`docs/product/build-brief.md`](docs/product/build-brief.md),
[`docs/self-hosting.md`](docs/self-hosting.md),
[`docs/performance.md`](docs/performance.md) and the UI rules in
[`docs/design-system.md`](docs/design-system.md).

## Development

```bash
go test ./...                 # backend and CLI tests
go vet ./...
go test -race ./...

pnpm --dir frontend check     # types and Svelte diagnostics
pnpm --dir frontend test      # Vitest unit/component tests
pnpm --dir frontend test:e2e  # Chromium, Firefox and WebKit smoke suite
pnpm --dir frontend build     # production assets for `-tags embed`
```

The Playwright suite starts a token-enabled API and Vite server automatically.
Install browsers once with
`pnpm --dir frontend exec playwright install chromium firefox webkit`.

```bash
./scripts/bench.sh            # startup, RSS, idle and latency measurements
docker compose up --build     # full container smoke
```

`scripts/bench.sh` and `scripts/ci-smoke.sh` expect `frontend/dist` from
`pnpm --dir frontend build` and build the server with `-tags embed`.

Working rules, the repo map and the docs index live in [`AGENTS.md`](AGENTS.md).

## Roadmap

- [x] **M1** — Go, SQLite, REST, project/epic/issue CRUD and the minimal workspace
- [x] **M1.1** — status/priority polish, Markdown descriptions, motion and icons
- [x] **M2** — board, comments, CLI, agent instructions, SSE and tests
- [x] **M2 editor** — CodeMirror Markdown editing, code blocks and copy actions
- [ ] **M3** — Docker, one binary, CI, releases, OpenAPI, benchmarks and self-hosting docs (implementation complete; first tagged release pending)
- [ ] **v0.2** — read-only GitHub repository, branch and pull request observation

## Contributing

Issues and pull requests are welcome. Open an issue first for larger changes, keep the surface small, and read [`AGENTS.md`](AGENTS.md) before sending a PR.

## License

Tablon is licensed under the **GNU Affero General Public License v3.0 only** (`AGPL-3.0-only`). See [`LICENSE`](LICENSE).

Commercial use is permitted under the AGPL. If you run a modified version as a network service, you must offer users the Corresponding Source as described by the license.

Copyright (C) 2026 Pablo Gutierrez.
