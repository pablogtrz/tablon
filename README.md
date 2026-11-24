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
  <a href="agent/Tablon.md"><img alt="AI-first: agents" src="https://img.shields.io/badge/AI--first-agents-67D4F5?style=flat-square&labelColor=0A0A0B"></a>
  <a href="docs/product/build-brief.md"><img alt="Self-hosted" src="https://img.shields.io/badge/self--hosted-yes-6EE7B0?style=flat-square&labelColor=0A0A0B"></a>
  <a href="docs/product/build-brief.md"><img alt="Status: v0.1 in development" src="https://img.shields.io/badge/status-v0.1%20in%20development-F5C451?style=flat-square&labelColor=0A0A0B"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#components">Components</a> ·
  <a href="#performance">Performance</a> ·
  <a href="#features">Features</a> ·
  <a href="#cli">CLI</a> ·
  <a href="#agent">Agent</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#development">Development</a> ·
  <a href="#license">License</a>
</p>

---

**Tablon** is a self-hosted issue tracker for small teams and coding agents. It keeps the workflow clear and the footprint small: a Go + SQLite backend, a calm dark-only Svelte UI, REST + SSE, and a `tablon` CLI built for token-efficient automation.

> Linear-like clarity. Self-hosted simplicity. Agent-ready by default.

Tablon ships as three parts that work together: the platform, the `tablon` CLI, and a portable agent adapter.

## Components

| Part | Role | Home |
|---|---|---|
| Platform | Go + SQLite server and the Svelte web UI; the source of truth | `cmd/server/`, `frontend/` |
| CLI | Token-efficient interface for agents, scripts and humans | [`cli/`](cli/README.md) |
| Agent | Portable adapter that teaches a coding agent to use the CLI | [`agent/Tablon.md`](agent/Tablon.md) |

## Why Tablon?

- **Small by design.** One Go service, SQLite in WAL mode, and no accounts, organizations or role system.
- **Agent native.** Agents discover, claim, update and comment through a compact CLI or the REST API.
- **Live without ceremony.** Per-project SSE keeps boards, issues and comments in sync; reconnect and refetch when needed.
- **Readable work.** Markdown descriptions and comments with a CodeMirror editor, headings, code blocks and copyable previews.
- **Fast to scan.** Dense issue tables, a four-column board, status at a glance and no decorative noise.
- **Dark only.** A deliberate dark interface with amber accents, accessible focus and reduced-motion support.

## Performance

Tablon keeps a small footprint and a small token bill. Measured on a shared
Raspberry Pi 4 (3.7 GiB RAM, Debian 13, kernel 6.18.50, aarch64, Docker 29.10.0),
seven runs per configuration. Workload: 1 project, 100 issues, 50 comments; idle
timeout `3s`. Full method and caveats live in
[`docs/performance.md`](docs/performance.md).

### Resources

Tablon is one Go process with SQLite and no Node runtime. It peaks at about
15 MiB RSS and uses about 4 MiB of working storage. Docker adds startup time and
image size but matches native steady-state RSS and request latency. Idle mode
does not release RSS, but it checkpoints the WAL and drops working storage to
about 0.2 MiB.

```text
Startup to /ready (s)
Native  █ 0.16
Docker  ████████ 1.30

On-disk footprint (MiB)
Native binary  ███████████ 11.2
Docker image   ██████████████████████████████ 30.4

SQLite working storage (MiB)
After workload  █████████████████████ 4.1
After idle      █ 0.2
```

| Metric | Native | Docker |
|---|---:|---:|
| Peak RSS | 14.3-14.5 MiB | 14.3-14.5 MiB |
| RSS after idle | 14.7-14.9 MiB | 14.7-14.9 MiB |
| First request after idle | 0.058-0.060 s | 0.059-0.061 s |
| List 50 issues | 0.057-0.062 s | 0.058-0.060 s |
| Binary / image | 11.2 MiB | 30.4 MiB |
| Data after workload | 4.1 MiB | 4.1-4.2 MiB |
| Data after idle | 0.2 MiB | 0.2 MiB |

A separate production container (dev image) on the same Pi measured **14.0 MiB
RSS after 1.85 h idle**. RSS is Resident Set Size: the physical pages currently resident for the
process, including the Go heap, stacks, binary text and shared libraries. It is
not virtual memory and not the live heap size. Go retains allocated heap pages,
and idle mode deliberately avoids `debug.FreeOSMemory`, so the idle/peak
difference is sampling noise rather than a memory increase.

#### How it compares

The closest self-hosted comparisons are Gitea and Vikunja (Go + SQLite);
Redmine is the traditional Ruby on Rails option. Feature sets differ, so this
compares runtime footprint, not feature parity. Competitor figures are planning
estimates, not measurements on this hardware.

| Product | Runtime | SQLite | Idle RAM | Peak RAM |
|---|---|---:|---:|---:|
| Tablon | Go | yes | ~14 MiB (measured) | ~14.5 MiB (measured) |
| Gitea | Go | yes | 40-120 MiB* | 80-220 MiB* |
| Vikunja | Go | yes | 50-130 MiB* | 100-250 MiB* |
| Redmine | Ruby on Rails | yes | 250-600 MiB* | 500-1,200 MiB* |

```text
Idle RAM, approximate midpoints (MiB)
Tablon   █ 14 (measured)
Gitea    ████████ 80
Vikunja  █████████ 90
Redmine  ███████████████████████████████████████████ 425
```

\* Planning estimates from a Gemini Deep Research summary of Gitea, Vikunja
and Redmine public documentation and community reports (2026-11-24). They are
not benchmarked on this Raspberry Pi or the same dataset. Vikunja's own stress
tests report roughly 199-439 MiB RSS under 500-3,000 simulated users, so treat
its peak range as optimistic for high concurrency.

### Token usage

The compact CLI is the intended agent path. The same five-step issue flow costs
about **5.8x fewer total tokens** than an external agent using raw REST because
`--compact` returns one line per operation instead of the full JSON object.
Counts use `tiktoken` `cl100k_base`; system/task prompts and network latency are
excluded.

```text
Total tokens (five-step flow)
External agent / REST  █████████████████████ 10,427
Tablon adapter / CLI   ████ 1,805

Workflow tokens
External agent / REST  █████████ 1,830
Tablon adapter / CLI   ██ 440
```

| Scenario | Context | Workflow | Total |
|---|---:|---:|---:|
| External agent / REST | 8,597 | 1,830 | 10,427 |
| Tablon adapter / CLI | 1,365 | 440 | 1,805 |

The CLI path is 5.8x fewer total tokens and 4.2x fewer workflow tokens. Use
`--json` when an agent needs exact fields; `--compact` is for routine status
work.

Reproduce with `./scripts/bench-runtime.sh` on Linux and
`python3 scripts/bench-tokens.py`.

## Features

| Area | What you get |
|---|---|
| Projects and epics | Project identifiers, statuses, gradient/emoji logos and epic tracking |
| Issues | `PROJ-123` identifiers, statuses, priorities, epics, agent assignment and filters |
| Board | Todo, In Progress, In Review and Done columns with drag, keyboard and menu moves |
| Comments | Chronological Markdown comments with edit/delete and agent attribution |
| Editor | CodeMirror 6 with live Markdown styling, fenced code blocks and copy actions |
| Automation | [`tablon` CLI](cli/README.md), JSON/compact output, idempotent creates, claim/release |
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

Open **http://127.0.0.1:8448**. Compose publishes only on loopback by default and
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

Open **http://127.0.0.1:8448**. Install the CLI from the same release or with
`go install github.com/pablogtrz/tablon/cli/tablon@latest`.

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
server on `:8448`; `go run` without `-tags embed` serves the API only.

The server embeds the built SPA when compiled with `-tags embed`, so production
and Docker need no Node runtime. The Vite dev server is only for frontend
development:

```bash
pnpm --dir frontend build
go build -tags embed -o bin/tablond ./cmd/server
```

See [`docs/self-hosting.md`](docs/self-hosting.md) for reverse proxies, token
guidance, backups and upgrades.

## CLI

```bash
go install github.com/pablogtrz/tablon/cli/tablon@latest

export TABLON_CLI_URL="http://127.0.0.1:8448"
export TABLON_CLI_TOKEN="optional-token"

tablon version
tablon --help
```

| Variable | Default | Purpose |
|---|---|---|
| `TABLON_CLI_URL` | `http://127.0.0.1:8448` | CLI API base URL |
| `TABLON_CLI_TOKEN` | empty | CLI bearer token |

| Task | Command |
|---|---|
| Projects | `tablon project list\|get\|create\|update` |
| Epics | `tablon epic list --project <P>\|get\|create\|update` |
| Issues | `tablon issue list --project <P>\|get\|create\|update` |
| Comments | `tablon issue comments\|comment`, `tablon comment update\|delete` |
| Assignment | `tablon issue claim\|release --agent <ID>` |

```bash
tablon project list --compact
tablon issue list --project TABLON --status todo --compact
tablon issue claim TABLON-17 --agent my-agent
tablon issue update TABLON-17 --status in-progress
tablon issue comment TABLON-17 --agent my-agent --body "Implementation started"
tablon issue release TABLON-17 --agent my-agent
```

- Use `--compact` for routine reads and `--json` when parsing. Never parse default human output.
- `issue create` and `issue comment` accept `--idempotency-key` for safe retries.
- Exit codes: `0` success, `1` runtime/API/network, `2` usage/config.

The full reference lives in [`cli/README.md`](cli/README.md).

## Agent

[`agent/Tablon.md`](agent/Tablon.md) is a portable, style-free adapter between a coding agent and the `tablon` CLI. It defines the connection, commands, output contract and guardrails only. The host repository or primary agent defines workflow: when to claim or update, branch and PR rules, status transitions and comment style.

| Client | Install |
|---|---|
| Claude Code | `~/.claude/agents/Tablon.md` or `.claude/agents/Tablon.md` |
| OpenCode | `~/.config/opencode/agents/Tablon.md` or `.opencode/agents/Tablon.md` |
| Codex | `~/.codex/agents/tablon.toml` or `.codex/agents/tablon.toml` |

Install details and fallbacks are in [`agent/README.md`](agent/README.md).

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `TABLON_HOST` | `127.0.0.1` | API bind address |
| `TABLON_PORT` | `8448` | API port |
| `TABLON_DATA_DIR` | `./data` | SQLite database directory |
| `TABLON_API_TOKEN` | empty | Optional bearer token for `/api/v1` |
| `TABLON_IDLE_ENABLED` | `true` | Enable the idle activity gate |
| `TABLON_IDLE_TIMEOUT` | `5m` | Quiet period before idle mode |
| `TABLON_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `TABLON_HOST_PORT` | `8448` | Compose-only host port published on `127.0.0.1` |

Native binds default to loopback. If you expose Tablon beyond your machine, set
`TABLON_API_TOKEN` and use HTTPS through a reverse proxy. The API contract is
served at `/openapi.yaml`.

## Architecture

| Layer | Choice |
|---|---|
| Backend | Go 1.27, `net/http`, pure-Go SQLite, embedded migrations and SPA |
| Frontend | Svelte 5, TypeScript, Vite, Tailwind CSS v4, CodeMirror 6 |
| Contract | REST + SSE, OpenAPI 3.0.3, structured JSON errors, cursor pagination |
| CLI | Go, REST only, in `cli/`, no direct database access |
| Agent | Portable style-free adapter in `agent/Tablon.md` |
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
go run ./cli/tablon --help    # CLI from source

pnpm --dir frontend check     # types and Svelte diagnostics
pnpm --dir frontend test      # Vitest unit/component tests
pnpm --dir frontend test:e2e  # Chromium, Firefox and WebKit smoke suite
pnpm --dir frontend build     # production assets for `-tags embed`
```

The Playwright suite starts a token-enabled API and Vite server automatically.
Install browsers once with
`pnpm --dir frontend exec playwright install chromium firefox webkit`.

```bash
./scripts/bench.sh            # macOS native startup, RSS, idle and latency
./scripts/bench-runtime.sh    # native vs Docker runtime and storage (Linux)
python3 scripts/bench-tokens.py --url <server> --cli <tablon>  # token cost
docker compose up --build     # full container smoke
```

`scripts/bench.sh` and `scripts/ci-smoke.sh` expect `frontend/dist` from
`pnpm --dir frontend build` and build the server with `-tags embed`.

Working rules, the repo map and the docs index live in [`AGENTS.md`](AGENTS.md).

## Roadmap

- [x] **M1** — Go, SQLite, REST, project/epic/issue CRUD and the minimal workspace
- [x] **M1.1** — status/priority polish, Markdown descriptions, motion and icons
- [x] **M2** — board, comments, CLI, agent adapter, SSE and tests
- [x] **M2 editor** — CodeMirror Markdown editing, code blocks and copy actions
- [ ] **M3** — Docker, one binary, CI, releases, OpenAPI, benchmarks and self-hosting docs (implementation complete; first tagged release pending)
- [ ] **v0.2** — read-only GitHub repository, branch and pull request observation

## Contributing

Issues and pull requests are welcome. Open an issue first for larger changes, keep the surface small, and read [`AGENTS.md`](AGENTS.md) before sending a PR.

## License

Tablon is licensed under the **GNU Affero General Public License v3.0 only** (`AGPL-3.0-only`). See [`LICENSE`](LICENSE).

Commercial use is permitted under the AGPL. If you run a modified version as a network service, you must offer users the Corresponding Source as described by the license.

Copyright (C) 2026 Pablo Gutierrez.
