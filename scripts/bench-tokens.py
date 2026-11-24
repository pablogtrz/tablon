#!/usr/bin/env python3
"""Token benchmark for the Tablon API and CLI.

Runs the same five-step issue workflow three ways and counts tokens with
tiktoken. Requires: python3 with tiktoken, a running Tablon server and a built
`tablon` CLI binary.

  python3 scripts/bench-tokens.py --url http://127.0.0.1:8448 --cli tmp/tablon-bench

The report is written to tmp/curl-vs-cli-tokens.md by default. Raw transcripts
are written to tmp/token-bench/ for audit.
"""

from __future__ import annotations

import argparse
import json
import os
import shlex
import subprocess
import sys
import urllib.error
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

TITLE = "Token benchmark: cache invalidation on project rename"
DESCRIPTION = """The project rename flow currently leaves cached issue lists pointing at the old identifier. Reproduce it by opening a project with at least two issues, renaming the project from the settings dialog, and returning to the board: the board still shows the previous identifier until a full reload. The API returns the new identifier immediately.

Expected behavior:
- The board, issue table and open issue route update without a reload.
- The active project cache is invalidated for both the identifier and the UUID.
- Existing issue identifiers keep their numeric suffix after the project identifier changes.
- The SSE stream reconnects with the new project reference when the rename completes.

Scope:
- Frontend cache invalidation and route parameter handling.
- No database migration.
- No changes to the REST payload shape.

Acceptance:
- A rename updates the sidebar, board header and deep link target.
- A hard reload and an in-app navigation show the same data.
- Unit tests cover the cache key transition."""
COMMENT = "Implemented and verified. Unit and integration tests pass; no regressions found."
AGENT = "token-bench"
PROJECT_IDS = {"external_api": "TOK1", "agent_api": "TOK2", "agent_cli": "TOK3"}


def token_count(enc, text: str) -> int:
    return len(enc.encode(text, disallowed_special=()))


def compact_json(value) -> str:
    return json.dumps(value, separators=(",", ":"), ensure_ascii=False)


class Bench:
    def __init__(
        self, base_url: str, cli: Path, tokenizer: str, out_dir: Path,
        include_agent_rest: bool = False,
    ):
        self.base_url = base_url.rstrip("/")
        self.cli = cli
        self.tokenizer = tokenizer
        self.out_dir = out_dir
        self.include_agent_rest = include_agent_rest
        self.out_dir.mkdir(parents=True, exist_ok=True)
        try:
            import tiktoken
        except ImportError as exc:
            raise SystemExit("tiktoken is required: python3 -m pip install tiktoken") from exc
        self.enc = tiktoken.get_encoding(tokenizer)
        self.env = os.environ.copy()
        self.env["TABLON_CLI_URL"] = self.base_url
        self.env.pop("TABLON_CLI_TOKEN", None)
        self.results = {}

    def http(self, method: str, path: str, body=None):
        data = compact_json(body).encode() if body is not None else None
        request = urllib.request.Request(
            self.base_url + path, data=data, method=method,
            headers={"Content-Type": "application/json", "Accept": "application/json"},
        )
        try:
            with urllib.request.urlopen(request) as response:
                raw = response.read().decode()
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode()
            raise RuntimeError(f"{method} {path} -> {exc.code}: {raw}") from exc
        request_text = f"{method} {path}\n"
        if body is not None:
            request_text += compact_json(body) + "\n"
        response_text = raw.rstrip("\n") + "\n"
        payload = json.loads(raw) if raw else None
        return request_text, response_text, payload

    def cli_run(self, args: list[str]):
        argv = [str(self.cli)] + args
        proc = subprocess.run(argv, capture_output=True, text=True, env=self.env)
        if proc.returncode != 0:
            raise RuntimeError(
                f"tablon {' '.join(args)} -> {proc.returncode}: {proc.stderr.strip()}"
            )
        command_text = shlex.join(["tablon"] + args) + "\n"
        response_text = proc.stdout
        if proc.stderr:
            response_text += proc.stderr
        response_text = response_text.rstrip("\n") + "\n"
        return command_text, response_text

    def ensure_clean_project(self, project_id: str, name: str) -> None:
        try:
            self.http("DELETE", f"/api/v1/projects/{project_id}")
        except RuntimeError:
            pass
        _, _, payload = self.http(
            "POST", "/api/v1/projects", {"name": name, "identifier": project_id}
        )
        if not payload or payload.get("identifier") != project_id:
            raise RuntimeError(f"project create returned {payload!r}")

    def cleanup_project(self, project_id: str) -> None:
        try:
            self.http("DELETE", f"/api/v1/projects/{project_id}")
        except RuntimeError:
            pass

    def api_workflow(self, project_id: str):
        steps = []
        req, res, issue = self.http(
            "POST",
            f"/api/v1/projects/{project_id}/issues",
            {"title": TITLE, "description": DESCRIPTION, "status": "todo"},
        )
        steps.append(("create issue", req, res))
        issue_ref = issue["identifier"]

        for status in ("in-progress", "in-review", "done"):
            req, res, _ = self.http(
                "PATCH",
                f"/api/v1/issues/{issue_ref}/status",
                {"status": status},
            )
            steps.append((f"status -> {status}", req, res))

        req, res, _ = self.http(
            "POST",
            f"/api/v1/issues/{issue_ref}/comments",
            {"content": COMMENT, "agent_id": AGENT},
        )
        steps.append(("comment", req, res))
        return steps

    def cli_workflow(self, project_id: str):
        steps = []
        req, res = self.cli_run([
            "issue", "create", "--project", project_id, "--title", TITLE,
            "--description", DESCRIPTION, "--status", "todo", "--compact",
        ])
        issue_ref = res.split(" | ", 1)[0].strip()
        steps.append(("create issue", req, res))

        for status in ("in-progress", "in-review", "done"):
            req, res = self.cli_run([
                "issue", "update", issue_ref, "--status", status, "--compact",
            ])
            steps.append((f"status -> {status}", req, res))

        req, res = self.cli_run([
            "issue", "comment", issue_ref, "--body", COMMENT,
            "--agent", AGENT, "--compact",
        ])
        steps.append(("comment", req, res))
        return steps

    def measure(self, scenario: str, context: str, steps):
        step_rows = []
        workflow_text = ""
        for name, req, res in steps:
            text = req + res
            workflow_text += text
            step_rows.append((name, token_count(self.enc, text), len(text)))
        context_tokens = token_count(self.enc, context)
        workflow_tokens = token_count(self.enc, workflow_text)
        self.results[scenario] = {
            "context_tokens": context_tokens,
            "workflow_tokens": workflow_tokens,
            "total_tokens": context_tokens + workflow_tokens,
            "steps": step_rows,
        }
        transcript = "## Context\n\n" + context + "\n\n## Workflow\n\n" + workflow_text
        (self.out_dir / f"{scenario}.txt").write_text(transcript)

    def run(self):
        openapi = (ROOT / "docs/api/openapi.yaml").read_text()
        adapter = (ROOT / "agent/Tablon.md").read_text()
        self.context_docs = {
            "openapi": (len(openapi), token_count(self.enc, openapi)),
            "adapter": (len(adapter), token_count(self.enc, adapter)),
        }

        self.ensure_clean_project(PROJECT_IDS["external_api"], "Token external API")
        self.measure(
            "external_api",
            openapi,
            self.api_workflow(PROJECT_IDS["external_api"]),
        )
        self.cleanup_project(PROJECT_IDS["external_api"])

        if self.include_agent_rest:
            self.ensure_clean_project(PROJECT_IDS["agent_api"], "Token agent API")
            self.measure(
                "agent_api",
                adapter + "\n\n" + openapi,
                self.api_workflow(PROJECT_IDS["agent_api"]),
            )
            self.cleanup_project(PROJECT_IDS["agent_api"])

        self.ensure_clean_project(PROJECT_IDS["agent_cli"], "Token agent CLI")
        self.measure(
            "agent_cli",
            adapter,
            self.cli_workflow(PROJECT_IDS["agent_cli"]),
        )
        self.cleanup_project(PROJECT_IDS["agent_cli"])

        self.write_report()

    def write_report(self):
        now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        labels = {
            "external_api": "External agent / REST",
            "agent_api": "Tablon adapter / REST",
            "agent_cli": "Tablon adapter / CLI",
        }
        scenarios = ["external_api", "agent_cli"]
        if self.include_agent_rest:
            scenarios.insert(1, "agent_api")
        lines = [
            "# Token benchmark — external API vs Tablon adapter",
            "",
            f"- Date: {now}",
            f"- Server: `{self.base_url}`",
            f"- Tokenizer: `{self.tokenizer}` (tiktoken)",
            f"- Description: {len(DESCRIPTION):,} characters",
            f"- Comment: {COMMENT!r}",
            f"- Workflow: create issue (todo) → in-progress → in-review → done → comment",
            "",
            "## Context",
            "",
            "| Document | Characters | Tokens |",
            "|---|---:|---:|",
            f"| `docs/api/openapi.yaml` | {self.context_docs['openapi'][0]:,} | {self.context_docs['openapi'][1]:,} |",
            f"| `agent/Tablon.md` | {self.context_docs['adapter'][0]:,} | {self.context_docs['adapter'][1]:,} |",
            "",
            "## Totals",
            "",
            "| Scenario | Context | Workflow | Total | vs CLI |",
            "|---|---:|---:|---:|---:|",
        ]
        cli_total = self.results["agent_cli"]["total_tokens"]
        for key in scenarios:
            r = self.results[key]
            ratio = r["total_tokens"] / cli_total if cli_total else 0
            lines.append(
                f"| {labels[key]} | {r['context_tokens']:,} | {r['workflow_tokens']:,} | "
                f"{r['total_tokens']:,} | {ratio:.1f}× |"
            )
        lines += [
            "",
            "## Workflow breakdown",
            "",
        ]
        headers = [labels[key] for key in scenarios]
        lines.append("| Step | " + " | ".join(headers) + " |")
        lines.append("|---|" + "---:|" * len(scenarios))
        step_names = [step[0] for step in self.results["external_api"]["steps"]]
        for index, name in enumerate(step_names):
            values = [self.results[key]["steps"][index][1] for key in scenarios]
            lines.append("| " + name + " | " + " | ".join(f"{v:,}" for v in values) + " |")
        lines += [
            "",
            "## Method",
            "",
            "Each scenario performs the same logical operations against the same server.",
            "The API scenario uses REST endpoints; the CLI scenario uses the `tablon`",
            "binary with `--compact` and `TABLON_CLI_URL` from the environment.",
            "",
            "- **Context** is the reference document an agent must load: the OpenAPI",
            "  contract for an agent that does not know the API, plus `agent/Tablon.md`",
            "  for adapter-guided scenarios.",
            "- **Workflow** counts the request/command text and the raw response/output",
            "  text for all five steps. It excludes the system prompt, the user task,",
            "  project setup/teardown and network latency.",
            "- Tokens are measured with `tiktoken` `cl100k_base` by default. Other",
            "  tokenizers will produce different absolute counts; compare scenarios,",
            "  not vendors.",
        ]
        if self.include_agent_rest:
            lines += [
                "- The adapter is CLI-oriented. The adapter/REST scenario is a deliberate",
                "  hybrid and therefore includes both the adapter and the OpenAPI contract.",
                "",
            ]
        lines += [
            "",
            "Raw transcripts: `tmp/token-bench/*.txt`.",
            "",
            "## Interpretation",
            "",
            f"- The CLI totals {self.results['external_api']['total_tokens'] / cli_total:.1f}× fewer tokens than the external REST scenario.",
            f"- Workflow-only: the CLI uses {self.results['external_api']['workflow_tokens'] / self.results['agent_cli']['workflow_tokens']:.1f}× fewer tokens "
            "than the REST workflow because `--compact` returns one line per operation instead of the full JSON object.",
            "- Most of the difference comes from output shape: REST returns descriptions, timestamps, UUIDs and every field on each write; `--compact` returns the identifier, status, priority and title.",
            "- For agents that need exact fields, `--json` is the right choice and is close to the raw API payload. `--compact` is for routine status work.",
            "",
        ]
        report = "\n".join(lines)
        (ROOT / "tmp" / "curl-vs-cli-tokens.md").write_text(report)
        (self.out_dir / "results.json").write_text(
            json.dumps(
                {
                    "date": now,
                    "server": self.base_url,
                    "tokenizer": self.tokenizer,
                    "context": self.context_docs,
                    "results": self.results,
                },
                indent=2,
            )
        )
        print(report)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--url", default="http://127.0.0.1:8448")
    parser.add_argument("--cli", default="tmp/tablon-bench", type=Path)
    parser.add_argument("--tokenizer", default="cl100k_base")
    parser.add_argument("--out-dir", default="tmp/token-bench", type=Path)
    parser.add_argument("--include-agent-rest", action="store_true")
    args = parser.parse_args()

    cli = args.cli if args.cli.is_absolute() else ROOT / args.cli
    if not cli.exists():
        raise SystemExit(
            f"missing CLI binary: {cli}\nBuild it with: go build -o {args.cli} ./cli/tablon"
        )
    out_dir = args.out_dir if args.out_dir.is_absolute() else ROOT / args.out_dir
    Bench(args.url, cli, args.tokenizer, out_dir, args.include_agent_rest).run()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit(130)
