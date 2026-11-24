# Tablon agent adapter

`Tablon.md` is the portable adapter between a coding agent and the `tablon` CLI. It defines the connection, commands, output contract and guardrails. It deliberately does not define workflow or style; the host repository or primary agent owns that.

## Install

| Client | Install | Notes |
|---|---|---|
| Claude Code | Copy or symlink `Tablon.md` to `~/.claude/agents/Tablon.md` or `.claude/agents/Tablon.md` | Uses the Markdown frontmatter. |
| OpenCode | Copy `Tablon.md` to `~/.config/opencode/agents/Tablon.md` or `.opencode/agents/Tablon.md` | `mode: subagent` is set in the frontmatter. |
| Codex | Copy `codex/tablon.toml` to `~/.codex/agents/tablon.toml` or `.codex/agents/tablon.toml` | Codex custom agents are TOML. The wrapper relies on `Tablon.md`, `tablon --help` or the served OpenAPI contract. |
| Other agents | Embed the Markdown body as instructions | No frontmatter is required. |

## Configure

```bash
export TABLON_CLI_URL="http://127.0.0.1:8448"
export TABLON_CLI_TOKEN="optional-token"
tablon version
```

Prefer environment variables. Do not commit tokens or place them in repository files.

## Scope

The adapter is intentionally style-free. Define in the host agent or repository:

- when to claim, update, comment or release;
- branch and pull-request conventions;
- status-transition rules;
- comment length, tone and required fields.
