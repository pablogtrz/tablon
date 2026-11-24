# `tablon` CLI

The `tablon` CLI is the token-efficient interface to a Tablon server for agents and scripts. It talks REST only and never opens the database.

## Install

```bash
go install github.com/pablogtrz/tablon/cli/tablon@latest
```

Release archives contain the same binary. From a source checkout:

```bash
go run ./cli/tablon --help
```

## Configure

| Setting | Default | Purpose |
|---|---|---|
| `TABLON_CLI_URL` | `http://127.0.0.1:8448` | API base URL |
| `TABLON_CLI_TOKEN` | empty | Optional bearer token |
| `--url` | env value | Per-command base URL override |
| `--token` | env value | Per-command token override |
| `--timeout` | `15s` | Request timeout |

```bash
export TABLON_CLI_URL="http://127.0.0.1:8448"
export TABLON_CLI_TOKEN="optional-token"
tablon version
```

The API contract is served at `$TABLON_CLI_URL/openapi.yaml`.

## Commands

### Projects

```bash
tablon project list [--include-archived]
tablon project get <project>
tablon project create --name TEXT [--identifier ID] [--description TEXT] [--status STATUS]
tablon project update <project> [--name TEXT] [--description TEXT] [--status STATUS]
```

`<project>` accepts a UUID or the project identifier. Statuses: `planning`, `active`, `archived`, `done`.

### Epics

```bash
tablon epic list --project <project>
tablon epic get <epic>
tablon epic create --project <project> --title TEXT [--description TEXT] [--status STATUS]
tablon epic update <epic> [--title TEXT] [--description TEXT] [--status STATUS]
```

`<epic>` is an epic UUID. Statuses: `backlog`, `in-progress`, `done`.

### Issues

```bash
tablon issue list --project <project> [--status STATUS] [--priority PRIORITY] \
  [--epic EPIC] [--q TEXT] [--sort FIELD] [--order asc|desc] [--limit N] [--cursor CURSOR]
tablon issue get <issue>
tablon issue create --project <project> --title TEXT [--description TEXT] [--status STATUS] \
  [--priority PRIORITY] [--epic EPIC] [--agent AGENT] [--idempotency-key KEY]
tablon issue update <issue> [--title TEXT] [--description TEXT] [--status STATUS] \
  [--priority PRIORITY] [--epic EPIC | --clear-epic]
```

`<issue>` accepts a UUID or an identifier such as `TABLON-17`. Statuses: `backlog`, `todo`, `in-progress`, `in-review`, `done`, `canceled`. Priorities: `urgent`, `high`, `medium`, `low`, `none`.

Sort fields: `created_at`, `updated_at`, `identifier`, `priority`. Pages default to 50 items, maximum 200.

### Comments

```bash
tablon issue comments <issue> [--limit N] [--cursor CURSOR]
tablon issue comment <issue> --body TEXT [--agent AGENT] [--idempotency-key KEY]
tablon comment update <comment> --body TEXT
tablon comment delete <comment>
```

`--body -` reads the comment body from stdin. `<comment>` is a comment UUID.

### Assignment

```bash
tablon issue claim <issue> --agent AGENT
tablon issue release <issue> --agent AGENT
```

`claim` sets the owner only when the issue is unclaimed and returns a conflict when another agent owns it. `release` requires the current owner. There is no direct reassign command; release then claim.

### Meta

```bash
tablon version
tablon --version
tablon help [group]
```

## Output

| Mode | Use |
|---|---|
| default | Human-readable tables and fields |
| `--compact` | Minimal one-line output for lists |
| `--json` | Raw API JSON |

`--json` and `--compact` are mutually exclusive. Global flags may appear before or after the command, but not after `--`.

```text
$ tablon issue list --project TABLON --status todo --compact
TABLON-17 | todo | high | Kanban board
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Runtime, API or network failure |
| `2` | Usage or configuration error |

Errors go to stderr.

## Retries

`issue create` and `issue comment` accept `--idempotency-key`. Reuse the same key and body when retrying the same write. The same key with a different body returns a conflict. Keys are 1–128 visible characters.

## Agent adapter

The portable, style-free adapter for coding agents lives in [`../agent/Tablon.md`](../agent/Tablon.md). Workflow policy belongs to the caller.
