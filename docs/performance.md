# Performance and memory

Measurements are engineering evidence, not targets. Re-run the commands below
on your own hardware before quoting numbers.

## Environment

- Date: 2026-11-24
- Machine: Intel Core i9-9880H, macOS 26.7.3 (Darwin x86_64)
- Go: 1.27.4
- Binary: `tablond` built with `-trimpath -ldflags "-s -w"` (11.9 MB)
- Data: 1 project, 100 issues, 50 comments for the process benchmark
- Native process benchmark; Docker/browser memory excluded

The repository `tmp/` directory is used for benchmark data. Some OS temp
filesystems added hundreds of milliseconds to SQLite startup on this machine;
using the repository filesystem keeps the numbers comparable.

## Process benchmark

Run:

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

The benefit of idle mode is the reusable activity gate for future background
workers (v0.2 GitHub polling) and predictable resource ownership, not a dramatic
memory win today.

## PRD targets

| Target | Measured | Result |
|---|---|---|
| Idle 10-25 MB | ~13 MB | Within target |
| Active, normal workload 30-70 MB | ~13 MB | Below target (target is an upper bound) |
| Normal workload under 100 MB | ~13 MB | Met |
| Native startup under 1 s | ~0.06 s | Met |

## Raspberry Pi 4 (arm64)

Environment: Raspberry Pi 4, 3.7 GiB RAM, Debian 13 kernel 6.18.50, aarch64,
Docker 29.10.0 and Compose v5.8.0. The native binary was extracted from the
arm64 image (11.2 MB stripped) and run directly three times with the same
workload. The container process was measured with `ps` because `docker stats`
reported `0B / 0B` on this cgroup setup.

| Metric | Native runs | Docker process |
|---|---|---|
| Startup to `/ready` | 0.160-0.162 s | n/a |
| RSS after start | 10.4-10.6 MB | n/a |
| RSS after workload | 14.5-14.6 MB | n/a |
| RSS after idle, state `idle` | 14.7-14.8 MB | 10.4 MB |
| First request after idle | 0.102-0.103 s | n/a |
| RSS after reactivation | 14.9-15.0 MB | 11.1 MB |
| List 50 issues latency | 0.102-0.103 s | n/a |

The Docker image is 31,872,202 bytes uncompressed. The image built successfully
on the Pi (Go and frontend builds inside Docker), the Compose service became
healthy with the real local database, persisted it across a restart, served the
SPA, deep links and `/openapi.yaml`, and ran as UID 10001 with a read-only root,
`cap_drop: ALL`, `no-new-privileges` and no Node. The imported database had 3
projects, 6 epics, 42 issues and 21 comments.

Idle mode again showed no material RSS reduction; the Pi numbers are consistent
with the Mac result.

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

The image size is recorded when the Docker build runs in CI. It is not included
in the native numbers above. The runtime adds Alpine, CA certificates and the
static `tablond` binary; the browser is excluded.

## Caveats

- Go GC and RSS are nondeterministic; treat small differences as noise.
- `debug.FreeOSMemory` is not used; see the idle result.
- API benchmarks run in-process and exclude TLS, proxies and real network RTT.
- Docker daemon and container memory are excluded from native RSS.
- Re-run after changing SQLite pragmas, indexes, payload sizes or the frontend
  payload. Update this file rather than claiming an unmeasured target.
