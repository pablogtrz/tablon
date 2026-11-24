# Performance and memory

Measurements are engineering evidence, not targets. Re-run the commands below
on your own hardware before quoting numbers.

## Environment

- Runtime benchmark: Raspberry Pi 4 Model B, 3.7 GiB RAM, Debian 13, kernel
  6.18.50+rpt-rpi-v8, aarch64, Docker 29.10.0 and Compose v5.8.0.
- Runtime image: `tablon:bench`, built on the Pi from the working tree.
- Runtime workload: 1 project, 100 issues, 50 comments; idle timeout `3s`;
  seven runs per configuration.
- Shared host: the Pi also ran unrelated services (production Tablon on
  `:8448`, Pi-hole, Portainer, Node-RED). These are real-world numbers, not an
  isolated lab.
- Source checkout: 1.2 MiB on disk, excluding `.git`, `tmp/` and
  `frontend/dist/`.
- macOS process benchmark: Intel Core i9-9880H, macOS 26.7.3 (Darwin x86_64),
  Go 1.27.4.
- Token benchmark: tiktoken `cl100k_base`; same five-step workflow against the
  same server for every scenario.

The repository `tmp/` directory is used for benchmark data. Some OS temp
filesystems added hundreds of milliseconds to SQLite startup on this machine;
using the repository filesystem keeps the numbers comparable.

## Raspberry Pi 4: native vs Docker

```bash
docker build -t tablon:bench .
./scripts/bench-runtime.sh
```

The script extracts `tablond` from the image for the native run, so both setups
use the same server binary. The native run is the long-running process a
pm2/systemd unit would manage; the Docker run uses the image entrypoint. It
starts each server with `TABLON_IDLE_TIMEOUT=3s`, creates one project, 100 issues
and 50 comments, then waits 5 seconds for idle.
RSS is read with `ps -o rss=` for both the native process and the container's
host PID; peak RSS is sampled after each workload request. Image size comes from
`docker image inspect`; storage is `du -sk` of the data directory or volume.
Docker startup includes container creation; native startup is process start only.

Seven runs:

| Metric | Native | Docker |
|---|---:|---:|
| Startup to `/ready` | 0.159-0.175 s | 1.15-1.71 s |
| RSS after start | 10.3-10.7 MiB | 10.3-10.6 MiB |
| RSS after first issue | 10.7-11.4 MiB | 10.8-11.2 MiB |
| Peak RSS during workload | 14.3-14.5 MiB | 14.3-14.5 MiB |
| RSS after workload | 14.8-14.9 MiB | 14.7-14.9 MiB |
| RSS after idle, state `idle` | 14.7-14.9 MiB | 14.7-14.9 MiB |
| First request after idle | 0.058-0.060 s | 0.059-0.061 s |
| RSS after reactivation, state `active` | 15.0-15.1 MiB | 14.9-15.1 MiB* |
| List 50 issues latency | 0.057-0.062 s | 0.058-0.060 s |
| Data / volume after workload | 4.1 MiB | 4.1-4.2 MiB |
| Data / volume after idle | 0.2 MiB | 0.2 MiB |
| Data / volume after shutdown | 0.2 MiB | n/a |

*One Docker run reported 15.7 MiB after reactivation; the other runs were 14.9-15.1 MiB.

### Storage

| Artifact | Size |
|---|---:|
| Source checkout (on disk, no `.git`/`tmp`/`dist`) | 1.2 MiB |
| Native binary | 11,731,104 B (11.2 MiB) |
| Docker image | 31,872,531 B (30.4 MiB) |
| SQLite database after workload (WAL active) | 4.1 MiB |
| SQLite database after idle | 192,512 B (0.2 MiB) |
| SQLite database after shutdown | 192,512 B (0.2 MiB) |

The Docker image contains Alpine 3.22, CA certificates and the static `tablond`
binary. It does not contain the `tablon` CLI, which is distributed separately.
The native setup avoids the image overhead but needs a Go/Node toolchain to
build and a checked-out source tree to produce the binary.

### Findings

- Steady-state RSS is the same within noise: roughly 14.8 MiB after the
  workload and 14.8 MiB while idle.
- Docker adds about 1.1-1.6 s to startup and about 19 MiB to the on-disk
  footprint. Request latency after startup is unchanged within noise.
- Idle mode does not materially reduce RSS. It does checkpoint/truncate the
  SQLite WAL: working storage falls from about 4.1 MiB to about 0.2 MiB after
  the idle timeout.
- Idle RSS is occasionally slightly higher than the sampled peak because peak
  RSS is sampled after each request, not continuously, and Go retains allocated
  heap pages. The difference is within noise, not a memory increase.
- A separate production container (dev image) on the same Pi measured 14.0 MiB
  RSS after 1.85 hours idle.
- The runtime is well below the 100 MiB normal-workload ceiling on both setups.

## Token benchmark

```bash
go build -o tmp/tablon-bench ./cli/tablon
python3 scripts/bench-tokens.py --url http://192.168.1.138:28800 --cli tmp/tablon-bench
```

Requires `tiktoken` (`python3 -m pip install tiktoken`). The script creates two
projects (`TOK1`, `TOK3`), runs the same flow against each, deletes them, and
writes `tmp/curl-vs-cli-tokens.md` plus raw transcripts in `tmp/token-bench/`.

Flow: create an issue with a 1,010-character description in `todo`, move it to
`in-progress`, `in-review` and `done`, then add a comment. The same description
and comment are used in every scenario.

Context:

- **External agent / REST:** an agent that does not know the API loads the full
  `docs/api/openapi.yaml`.
- **Tablon adapter / CLI:** loads `agent/Tablon.md` and uses the `tablon` binary
  with `--compact`.

| Document | Characters | Tokens |
|---|---:|---:|
| `docs/api/openapi.yaml` | 37,844 | 8,597 |
| `agent/Tablon.md` | 5,306 | 1,365 |

| Scenario | Context | Workflow | Total | vs CLI |
|---|---:|---:|---:|---:|
| External agent / REST | 8,597 | 1,830 | 10,427 | 5.8× |
| Tablon adapter / CLI | 1,365 | 440 | 1,805 | 1.0× |

| Step | External agent / REST | Tablon adapter / CLI |
|---|---:|---:|
| Create issue | 583 | 242 |
| Status to in-progress | 365 | 36 |
| Status to in-review | 365 | 36 |
| Status to done | 363 | 34 |
| Comment | 154 | 92 |

The CLI path uses **5.8× fewer total tokens** than the external REST baseline
and **4.2× fewer workflow tokens**. Most of the difference is output shape: REST
returns descriptions, timestamps, UUIDs and every field on each write;
`--compact` returns the identifier, status, priority and title. For exact
fields, `--json` is the right choice and is close to the raw API payload.

Caveats:

- tiktoken is not Claude's tokenizer. Absolute counts vary by model; compare
  scenarios, not vendors.
- System prompts, the user task, project setup/teardown and network latency are
  excluded.
- API requests are modeled as method, path and compact JSON body; CLI commands
  include the full shell command line and `--compact` output.
- Token counts are text tokens, not billing units.

## PRD targets

| Target | Measured | Result |
|---|---|---|
| Idle 10-25 MB | 13-15 MB | Within target |
| Active, normal workload 30-70 MB | 13-15 MB peak | Below target (target is an upper bound) |
| Normal workload under 100 MB | 13-15 MB peak | Met |
| Native startup under 1 s | 0.06 s macOS, 0.16 s Pi | Met |

## macOS native process benchmark

```bash
./scripts/bench.sh
```

Run `pnpm --dir frontend build` first: the script compiles `tablond` with
`-tags embed`, warms the binary with a `--version` execution, starts it with
`TABLON_IDLE_TIMEOUT=3s`, creates a small workload, and records startup, RSS and
first-request latency. It uses Go, curl, perl, python3, ps and awk.

Representative run (three runs on 2026-11-24 produced the same shape):

| Metric | Value |
|---|---|
| Startup to `/ready` | 0.06-0.08 s |
| RSS after start | 9.0-9.3 MB |
| RSS after workload | 12.8-12.9 MB |
| RSS after idle, state `idle` | 13.0-14.0 MB |
| First request after idle | 0.030-0.036 s |
| RSS after reactivation | 13.1-14.1 MB |
| List 50 issues latency | 0.029-0.036 s |

### Idle mode result

Idle mode does **not** materially reduce RSS at this workload. The server keeps
the HTTP listener and SQLite open; `PRAGMA shrink_memory` and releasing the idle
writer connection are best-effort. A forced `debug.FreeOSMemory()` was measured
and removed because it increased RSS on macOS (about +0.8-1.4 MB) without a
useful reduction. The measured idle/active difference is within noise.

Idle mode still truncates the SQLite WAL, which reduces on-disk working storage
from about 4.1 MiB to about 0.2 MiB on the Pi workload. The reusable activity
gate remains useful for future background workers (v0.2 GitHub polling).

## Go benchmarks

Store benchmarks seed about 10k issues and 5k comments outside the timed loop.
API benchmarks use an in-process `httptest` server with auth disabled.

```bash
go test -run '^$' -bench . -benchmem -count 5 ./internal/store ./internal/api
```

Representative single run (`-benchtime=50x`):

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| List 50 of 10k issues | 203,586 | 47,157 | 1,172 |
| List filtered status | 346,349 | 47,640 | 1,176 |
| List next cursor | 214,477 | 47,978 | 1,183 |
| Get issue by identifier | 47,343 | 1,984 | 61 |
| Create issue | 297,754 | 3,242 | 65 |
| Update issue status | 152,227 | 748 | 17 |
| Claim issue (claim + release) | 304,384 | 4,880 | 150 |
| Claim issue, contended | 140,489 | 3,198 | 89 |
| List 50 of 5k comments | 123,985 | 30,086 | 800 |
| Get project | 38,800 | 1,712 | 51 |
| API list issues | 559,096 | 117,886 | 1,393 |
| API get issue | 186,379 | 10,965 | 170 |
| API get project | 180,580 | 10,992 | 161 |
| API create comment | 345,868 | 17,709 | 249 |
| API idempotent comment | 490,069 | 21,111 | 323 |
| API system status | 143,341 | 10,123 | 125 |

These are single-machine numbers; use `-count 5` and compare medians rather
than treating one run as a regression gate. CI only compiles and smoke-runs the
benchmarks (`-benchtime=1x`) so noisy hardware does not block pull requests.

## Query plans

`internal/store/query_plan_test.go` asserts that the main issue and comment
queries use `idx_issues_*` and `idx_comments_issue_id`. It guards against an
accidental full scan after a query change. It does not benchmark every filter.

## Docker image

The runtime image is 31,872,531 bytes (30.4 MiB) for linux/arm64. It adds Alpine,
CA certificates and the static `tablond` binary; the browser and the CLI are
excluded. The CLI is distributed separately.

## Caveats

- Go GC and RSS are nondeterministic; treat small differences as noise.
- `debug.FreeOSMemory` is not used; see the idle result.
- API benchmarks run in-process and exclude TLS, proxies and real network RTT.
- Docker daemon and container memory are excluded from native RSS.
- Docker startup includes container creation; compare steady state, not startup.
- Token counts use one tokenizer and one reference document snapshot. Re-run
  after changing the OpenAPI contract or `agent/Tablon.md`.
- Re-run after changing SQLite pragmas, indexes, payload sizes or the frontend
  payload. Update this file rather than claiming an unmeasured target.
