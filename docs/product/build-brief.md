# Tablon v0.1 — build brief

**Status:** recorded 2026-11-23. M1, M1.1, M2 and M3 implemented; first tagged release pending review/approval.
**Filter:** "REMEMBER THE MAIN GOAL IS TO HAVE A EXTREMELLY LIGHTWEIGHT PRODUCT MANAGEMENT TOOL" (user, 2026-11-23).

## Brief (user's words)

- "extremely lightweight" (`tmp/PROJECT.md`).
- "AI oriented ... consuming as few tokens as possible" (`tmp/PROJECT.md`).
- "Modern clean UI: Simple, just what you need, where it has to be. But nice looking." (`tmp/PROJECT.md`).
- "CLI has to be for agent to use the platform and communicate with the projects and issues. It IS NOT for managing the development of this project itself." (user, 2026-11-23).
- "Defer GitHub." (user, 2026-11-23).
- "No light theme." (user, 2026-11-23).

## Goal

A self-hosted, single-user issue tracker that a human runs with one binary or Docker Compose and an agent drives through REST and the `tablon` CLI, with a dark-only Linear-like UI. lightweight at runtime, dependency count and API surface.

## Surface and users

- Self-hosted Go binary and Docker container.
- Browser UI at desktop, tablet and mobile widths.
- `tablon` CLI for humans and agents.
- Agents via the REST API and CLI.
- Single user; no accounts.

## User requirements

- Dark only. No light theme. (user, 2026-11-23)
- Project logo: default gradient tile plus optional emoji. (user, 2026-11-23)
- Project identity: visual emoji picker, curated gradient tiles and a custom two-stop gradient maker. (user, 2026-11-25)
- Project statuses: Planning, Active, Archived, Done. (user, 2026-11-23)
- Issue slugs `in-progress` and `in-review`; epic statuses Backlog, In Progress, Done. (user, 2026-11-23)
- Comments can be edited and deleted. (user, 2026-11-23)
- Agents can create issues, epics and projects when specified. (user, 2026-11-23)
- Module `github.com/pablogtrz/tablon`; server `tablond`; CLI `tablon`. (user, 2026-11-23)
- No `PM_*` env prefix; different prefixes are fine. (user, 2026-11-23)
- GitHub integration deferred to v0.2. (user, 2026-11-23)
- Full adjusted v0.1 before a public demo, with an M1 checkpoint. No deadline. (user, 2026-11-23)
- Minimal `AGENTS.md` as a docs index; docs stay concise and load only when relevant. (user, 2026-11-23)
- No external React-based UI registry; no React, Radix, dnd-kit, TanStack or date-fns in the Svelte app. (user, 2026-11-23)
- Auth is optional; LAN use does not require it. (user, 2026-11-23)
- Minimal claim/release for agents. (user, 2026-11-23)

## Model proposals

- **Backend (model):** Go 1.27, stdlib `net/http` router, pure-Go SQLite (`modernc.org/sqlite`), embedded SQL migrations with a small runner, `log/slog`, embedded frontend.
- **Frontend (model):** Svelte 5 SPA, history routing with Go fallback, typed `fetch` + runes, native drag-and-drop with keyboard/status-menu fallback, Vitest + Playwright as dev-only dependencies.
- **Contract (user-approved, 2026-11-23):** UUIDv7 internal IDs; immutable project identifier and `PROJ-123` issue identifiers; per-project counter that survives deletes; project identifiers never reused; cursor pagination `(created_at, id)` with default 50 and max 200; last-write-wins PATCH; atomic claim/release; `Idempotency-Key` on POST issues and comments; UTC RFC3339 timestamps; error shape `{ "error": { "code", "message", "details" } }`.
- **Claim/release (model):** `POST /api/v1/issues/{id}/claim` sets `agent_id` only when unclaimed; `409` when taken; `POST /api/v1/issues/{id}/release` requires the matching agent; no TTL, queue or scheduler.
- **Auth (model):** `TABLON_API_TOKEN` is optional and empty by default. When set, it gates all `/api/v1` reads and writes; `/health` and `/ready` stay public. SPA uses an unlock screen and `sessionStorage`; SSE uses `fetch` streaming with `Authorization`. Native default binds `127.0.0.1`; Compose binds loopback by default. A non-loopback bind without a token logs one informational line, never blocks.
- **Project status tones (model):** Planning `--status-backlog`, Active `--status-progress`, Done `--status-done`, Archived `--text-3`.
- **Logo (model):** eight curated gradient pairs in `app.css`, selected deterministically from the project ID, plus an optional custom two-stop gradient (`logo_gradient_from`, `logo_gradient_to`, `logo_gradient_angle`); an emoji sits on top.
- **Milestones (model):**
  - M1: project/issue/epic CRUD, SQLite, REST, minimal UI (dashboard, list, detail), tests. Checkpoint.
  - M2: board, comments edit/delete, CLI, `AGENTS.md`, SSE.
  - M3: idle mode, Docker, CI, release.
  - GitHub sits after M3, in v0.2.
- **Repo (model):** `cmd/server`, `internal/`, `cli/`, `agent/`, `frontend/`, `migrations/`, `docs/`. `tmp/` remains gitignored and untracked. `.opencode/`, `.agents/`, `skills-lock.json` and `tmp/` are ignored; `opencode.json` is tracked.
- **Environment (model):** `TABLON_*` for the server, `TABLON_CLI_URL` and `TABLON_CLI_TOKEN` for the CLI, `GITHUB_*` when GitHub returns. Go 1.27.4 at `~/sdk/go`; Node 22 via `.nvmrc`.

## Out of scope

- Multi-user support, accounts, organizations, teams, roles, permissions, profiles.
- Light theme.
- GitHub integration, polling, webhooks and Git link tables in v0.1.
- Windows ARM64.
- Image-upload logos.
- Per-project GitHub repository configuration.
- Git branch, commit or pull request creation.
- MCP server.
- Sprints, time tracking, story points, resource allocation, advanced reporting, billing, SaaS.
- Distributed job scheduling, complex workflow automation, AI model execution.

## Constraints

- Go 1.27.4, Node 22.23.4, pnpm 11.2.2, SQLite WAL.
- License: AGPL-3.0-only.
- Pure-Go SQLite driver for CGO-free cross-platform builds.
- No heavy dependencies; no React, Radix, dnd-kit, TanStack or date-fns.
- Dark only.
- Docs stay short; `AGENTS.md` stays an index.
- No commits without explicit approval.

## Failure behavior

- Invalid core configuration fails closed with a clear error.
- Missing or invalid bearer token returns `401`; no secrets in logs or responses.
- Claim conflicts return `409`.
- API errors use the structured shape; agent retries are protected by idempotency keys on POST.
- SSE reconnect refetches state; no event replay in v0.1.
- Optional integrations fail open with warnings in system status when they return in v0.2.

## Taste references

- Linear for density and workflow clarity.
- Tablon "Obsidian Glass" design system for the visual language.

## Outside judge

Interviewers and open-source users. No named individual.

## Assumptions

- No deadline; the project is personal and built as a whole, with an M1 checkpoint.
- GitHub is a v0.2 feature.
- Auth is optional and off by default; LAN use is expected to be unauthenticated.
- Claim/release is cooperative and minimal; it is not a scheduler.
- Go is installed at `~/sdk/go` with `/usr/local/bin` wrappers; Node 22 is selected by `.nvmrc`.
- The existing `tmp/PRD.md` and `tmp/PROJECT.md` stay untracked for now.

## Approval scope

Decisions approved through the 2026-11-23 grill. M1 and M1.1 implemented and committed. The brief is open for corrections.

## Grilled

2026-11-23, three rounds, self-hosted web/CLI/agent surface.

## M1 checkpoint — 2026-11-23

Delivered:

- Go server with `net/http`, `log/slog`, graceful shutdown and optional bearer auth.
- SQLite in WAL mode with embedded migrations, foreign keys, busy timeout and a single writer connection.
- Projects, epics and issues: create, list, get, update, delete; project identifier registry; per-project issue numbering that survives deletes.
- Issue filters, priority/date/identifier sorting and cursor pagination; atomic claim/release; epic/project validation.
- Minimal UI: dashboard, project issues and epics, issue detail, project settings, create/edit/delete dialogs, settings page, mobile navigation, tablet rail and dark tokens.
- Backend tests for services, API, config and migrations; frontend type checks and production build.

Verified:

- `go test ./...`, `go vet ./...`, `go test -race ./...`, `pnpm check`, `pnpm build`.
- Browser smoke: create project, create issue, save status, create epic, mobile navigation, browser auth flow.
- 360, 390, 768, 1024 and 1440px checked with no horizontal overflow.

Debt and deferrals:

- Frontend unit tests now cover Markdown rendering and task toggles; a full component/E2E suite is still planned for M2.
- Comments, board, CLI, SSE and the distributable agent instructions remain in M2.
- Projects and epics are not paginated yet; issues are. Revisit if collections grow.


## M1.1 UI polish — 2026-11-23

Delivered:

- Issue table order `ID | Status | Priority | Issue | Epic | Agent`; tinted status cells with status icons; neutral priority badges with priority icons; mobile cards mirror the same order above the title.
- Issue-detail header no longer duplicates status/priority badges; properties use the rich selects.
- Native `appearance: base-select` selects for issue status, priority, epic, project status, epic status and filters, with plain native fallback in unsupported browsers. Issue-view properties save immediately with optimistic rollback.
- Markdown descriptions: `marked` + DOMPurify; https images, GFM tables, task lists, toolbar, Write/Preview, image/link dialogs and table insertion.
- Interactive task checkboxes: immediate optimistic PATCH in read mode; draft sync in preview; focus restoration.
- Independent title editing with blur/Enter save, empty-title revert and failure guard.
- Description read/edit modes with amber edit border, Save/Cancel, unsaved-change dialogs, in-app navigation guard and `beforeunload` guard.
- Subtle route, tab, dialog, drawer and select motion; route scroll-to-top and heading/region focus; reduced motion respected.
- Lucide icons for navigation and actions; pointer/not-allowed cursors.
- Issue view lazy-loaded as a separate chunk.

Verified:

- `pnpm check`, `pnpm test` (10 Markdown tests), `pnpm build`.
- Browser smoke: status/priority/epic immediate saves, title blur/Enter/empty, Markdown edit/save/cancel, toolbar image/table, preview checkboxes, read-mode checkbox persistence, in-app and browser-back unsaved guards, route focus and scroll.
- 360, 390, 768, 1024 and 1440px checked with no horizontal overflow.
- Initial JS: 170.68 kB raw / 55.61 kB gzip; issue chunk: 111.60 kB / 37.16 kB gzip; CSS: 41.76 kB / 9.39 kB gzip. The Markdown editor/parser is not in the initial chunk.

Deferred and known limits:

- Project and epic descriptions stay plain textareas.
- No image upload; only external https URLs.
- Firefox and Safari below 27 get the plain native select; no polyfill.
- Comments and board remain in M2.
- The temporary Playwright diagnostic dependency was removed after verification.


## M2 checkpoint — 2026-11-24

Delivered:

- Comments API with chronological pagination, edit/delete, cascade and agent ids.
- `Idempotency-Key` for POST issues and comments, with 24h replay storage and conflict handling.
- Four-column board (`Todo`, `In Progress`, `In Review`, `Done`) with native HTML5 drag, optimistic persistence, mobile snap scrolling and hidden backlog/canceled counts.
- Board keyboard movement (`Alt+ArrowLeft/Right`), a six-status move menu, focus restoration and polite announcements.
- Comments UI on issue detail: Markdown composer, chronological list, agent/local label, edited marker, inline edit, confirm delete and focus management.
- `tablon` CLI with projects, epics, issues, comments, claim/release, JSON and compact output, predictable exit codes and idempotent creates.
- Distributable `agent/Tablon.md` adapter, client wrappers and its docs-index entry.
- Per-project SSE stream with an in-process hub, heartbeat, reconnect/refetch, visibility handling and polite live regions.
- Vitest component/unit tests and Playwright smoke coverage for board, comments, unsaved guards, auth and responsive layout.

Verified:

- `go test ./...`, `go vet ./...`, `go test -race ./...`.
- `pnpm check`, `pnpm test` (23 tests), `pnpm build`.
- `pnpm test:e2e` (6 Chromium tests) against a token-enabled Go server.
- CLI smoke against a temporary server: project create, issue create/list/get, comment create/list, claim/release, JSON output.
- No horizontal page overflow at 360, 390, 768, 1024 and 1440px on dashboard, project, board and issue detail.
- Initial JS: 187.13 kB raw / 60.81 kB gzip; issue chunk: 122.11 kB / 40.01 kB gzip; CSS: 45.94 kB / 10.09 kB gzip.

Known limits:

- SSE has no event replay or `Last-Event-ID`; reconnect always refetches.
- Playwright runs Chromium only; the browser matrix and CI wiring remain in M3.
- Board loads at most 1,000 project issues; comments load in pages of 200 with a load-more control.
- Idempotent replays are resolved after resource validation; retrying after the target issue/project is deleted returns not found.
- No commits or PRs yet; review is pending.


## M2 editor expansion — 2026-11-24

Delivered:

- Replaced the Markdown textarea with a CodeMirror 6 editing surface in descriptions and comments.
- Added custom Markdown decorations for headings, bold/italic/strikethrough, inline code, links, list/task markers, quotes, rules and fenced code blocks. Markdown markers stay visible.
- Kept the existing toolbar, Write/Preview tabs, undo/redo, shortcuts, dirty state, unsaved guards and comment/description reuse.
- Added `CodeBlockDialog` with an optional language and a custom language input.
- Added language headers and keyboard-accessible Copy actions to rendered code blocks, with polite copy feedback and DOMPurify-safe DOM construction.
- Moved editor state behind a controller API (focus, selection, replace, prefix, insert) and added pure action/scan tests.

Verified:

- `pnpm check`, `pnpm test` (40 tests), `pnpm build`.
- `pnpm test:e2e` with 4 workers (8 Chromium tests), including rich Markdown editing, code block insertion/copy, comments, guards, SSE and layout.
- Initial JS: 187.53 kB raw / 60.98 kB gzip (materially unchanged from 187.13 / 60.81).
- IssueView chunk: 410.11 kB raw / 134.16 kB gzip, lazy-loaded only on issue detail.
- CSS: 49.08 kB raw / 10.58 kB gzip.

Known limits:

- The custom highlighter is regex/line based, not a full Markdown parser; unusual nested constructs may be styled conservatively.
- No split editor/preview pane yet.
- Fenced code blocks have language labels but no language-specific syntax highlighting.
- CodeMirror adds about 287 kB raw to the lazy IssueView chunk; the initial bundle is unchanged.


## M3 checkpoint — 2026-11-24

Delivered:

- One-binary server: `go:embed` SPA behind `-tags embed`, an empty `!embed` stub for development/tests, SPA history fallback, hashed asset caching, security headers/CSP, public `/openapi.yaml`, and a clear 503 when the SPA is absent.
- Idle mode: `TABLON_IDLE_ENABLED` and `TABLON_IDLE_TIMEOUT`, an activity gate around meaningful `/api/v1` traffic, in-flight tracking, SSE-aware behavior, `PRAGMA shrink_memory`, idle writer release, status fields and Settings rows.
- OpenAPI 3.0.3 contract in `docs/api/openapi.yaml`, embedded and served at `/openapi.yaml`, with route-table and schema contract tests using a test-only kin-openapi dependency.
- Docker: multi-stage Alpine image, non-root UID/GID 10001, read-only root, cap drop, `/data` named volume, `/health` healthcheck, OCI labels; Compose loopback defaults; `.env.example`; no Node runtime.
- CI: backend tests/vet/race/migrations/benchmark compile, frontend check/test/build, Playwright Chromium/Firefox/WebKit, embedded binary smoke, Docker smoke, five-target cross-builds and GoReleaser config check.
- Release: `internal/version` ldflags, `tablond --version`, `tablon version`/`--version`, GoReleaser bundle (server + CLI) and CLI-only archives, `SHA256SUMS`, grouped notes, multi-arch GHCR image with SBOM/provenance.
- Benchmarks: store/API benchmarks, query-plan assertions, `scripts/bench.sh`, `scripts/bench-runtime.sh`, `scripts/bench-tokens.py`, and recorded results in `docs/performance.md`.
- Self-hosting and release docs: `docs/self-hosting.md`, `docs/release.md`, README quick starts, AGENTS index, browser install and smoke instructions.

Verified:

- `go test ./...`, `go vet ./...`, `go test -race ./...`, `go test -tags embed ./...`.
- `pnpm check`, `pnpm test` (40 tests), `pnpm build`, and Playwright (Chromium 8 passed; Firefox/WebKit 7 passed + 1 Chromium-only skip each).
- `go build -tags embed`; `scripts/ci-smoke.sh` (embedded SPA, deep links, OpenAPI, auth, JSON 404s).
- `go test ./...` with `frontend/dist` removed (empty `!embed` stub), then embedded smoke again.
- `goreleaser check`, `goreleaser build --snapshot --clean`, and `goreleaser release --snapshot --clean` (10 archives + `SHA256SUMS`; versions injected).
- `docker compose config`.
- `scripts/bench.sh`: startup 0.06-0.08 s, RSS 9.0-9.3 MB at start, 12.8-12.9 MB after workload, 13.0-14.0 MB idle, first request after idle 0.030-0.036 s; binary 11.9 MB.
- Raspberry Pi 4 arm64: `tablon:bench` built on the Pi; `scripts/bench-runtime.sh` measured native vs Docker across seven runs. Native startup 0.159-0.175 s, Docker 1.15-1.71 s; peak RSS 14.3-14.5 MiB on both; data 4.1 MiB while working and 0.2 MiB after idle; binary 11.2 MiB, image 30.4 MiB. The five-step token benchmark measured 1,804 tokens for the adapter/CLI path versus 10,424 for external REST. Full data in `docs/performance.md`.

Known limits:

- Idle mode does not materially reduce RSS at this workload. It does checkpoint/truncate the SQLite WAL, reducing working storage from about 4.1 MiB to about 0.2 MiB. Forced `debug.FreeOSMemory` was removed after it increased RSS; the activity gate is the durable part.
- Release artifacts were verified with a snapshot only; the first tagged release and GHCR push remain pending review/approval.
- A server binary without `-tags embed` serves the API but not the SPA.
- Benchmark figures are from one machine; use `-count` and medians before treating them as gates.

## Review fixes — 2026-11-25

Addressed the P0 and P1 findings from the engineering review:

- Markdown sanitization now keeps only task checkboxes and language classes, so agent-authored content cannot inject form controls or arbitrary utility classes.
- `issue update --clear-epic=false` no longer clears the epic; the flag combination is covered by tests.
- `issue list` and `issue comments` print the next cursor in human and compact output; the CLI adapter and docs match.
- CLI API errors include the HTTP status and error code, for example `409 conflict: ...`.
- SSE closes on server shutdown through `RegisterOnShutdown`; the client exposes connecting, live, offline and token-required states, and stops retrying on 401/403.
- Project, issue and comment loads abort stale requests and ignore out-of-order responses; the shell shows one live indicator.
- Timestamps are stored fixed-width with migration `0003`, so TEXT ordering is chronological; legacy rows are normalized.
- Project and editor tabs implement the full ARIA tab contract; mobile comment/board actions meet the 40px target; dialogs expose descriptions and the issue detail is a labelled region.

Verified: `go test ./...`, `go vet ./...`, `go test -race ./...`, `go test -tags embed ./...`, `pnpm check`, `pnpm test` (50), `pnpm build`, `pnpm test:e2e` (37 passed, 2 skipped across Chromium, Firefox and WebKit).
