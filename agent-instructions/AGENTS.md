# Working with Tablon

Tablon is the source of truth for issues in this repository. Use the `tablon` CLI; never read or write the Tablon database directly.

## Setup

```bash
go install github.com/pablogtrz/tablon/cmd/tablon@latest

export TABLON_CLI_URL="http://127.0.0.1:8080"
export TABLON_CLI_TOKEN="optional-token"

tablon --help
tablon version
```

Prefer environment variables. Do not commit tokens or put them in repository files.
The machine-readable API contract is served at `$TABLON_CLI_URL/openapi.yaml`.

## Core loop

1. Discover work:
   ```bash
   tablon project list --compact
   tablon issue list --project <PROJECT> --status todo --compact
   ```
2. Choose an issue and claim it with a stable agent id:
   ```bash
   tablon issue claim <ISSUE> --agent <agent-id>
   ```
3. Read the full issue before changing code:
   ```bash
   tablon issue get <ISSUE>
   ```
4. Mark it in progress:
   ```bash
   tablon issue update <ISSUE> --status in-progress
   ```
5. Implement. Create a branch named `<IDENTIFIER>-<short-slug>`.
6. Report progress with short comments when useful:
   ```bash
   tablon issue comment <ISSUE> --agent <agent-id> --body "Implementation started"
   ```
7. When a pull request is open:
   ```bash
   tablon issue update <ISSUE> --status in-review
   ```
8. Release when your work is done:
   ```bash
   tablon issue release <ISSUE> --agent <agent-id>
   ```

## Commands

| Task | Command |
|---|---|
| Projects | `tablon project list\|get\|create\|update` |
| Epics | `tablon epic list --project <P>\|get\|create\|update` |
| Issues | `tablon issue list --project <P>\|get\|create\|update` |
| Comments | `tablon issue comments <ISSUE>\|issue comment <ISSUE> --body ...` |
| Edit/delete comments | `tablon comment update <ID> --body ...\|comment delete <ID>` |
| Claim/release | `tablon issue claim <ISSUE> --agent <ID>\|issue release <ISSUE> --agent <ID>` |

Run `tablon <command> -h` for flags.

## Output

- Use `--compact` for routine reads. It minimizes tokens:
  ```text
  TABLON-17 | in-progress | high | Kanban board
  ```
- Use `--json` only when you need to parse fields. Never parse human output.
- Errors go to stderr. Exit codes: `0` success, `1` runtime/API/network, `2` usage/config.

## Retries

Create commands accept `--idempotency-key`. Reuse the same key when retrying the same
create; a different body with the same key returns a conflict.

```bash
tablon issue create --project <P> --title "Fix login" --idempotency-key retry-1
tablon issue comment <ISSUE> --body "Retry-safe" --idempotency-key retry-1
```

## Branches and pull requests

- Branch: `<IDENTIFIER>-<short-slug>`, for example `TABLON-17-kanban-board`.
- PR title: `<IDENTIFIER>: <summary>`.
- Mention the issue identifier in commit messages and the PR body.
- Update Tablon status when the PR opens and when it merges.

## Rules

- Never access SQLite or the server's data directory.
- Never assume an issue is complete without verifying the implementation and tests.
- Keep comments short and factual: decisions, blockers, verification.
- Do not commit secrets. Redact tokens from logs.
- Ask before destructive actions such as deleting issues, epics or projects.
- Prefer `--agent` on comments and claims so progress is attributable.
