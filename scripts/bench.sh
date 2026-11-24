#!/usr/bin/env sh
# Reproducible native benchmark for startup, active/idle RSS and a small workload.
# Requires: go, curl, perl, python3, ps, awk.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
mkdir -p "$ROOT/tmp"

PORT=${TABLON_BENCH_PORT:-28790}
IDLE_TIMEOUT=${TABLON_BENCH_IDLE_TIMEOUT:-3s}
if curl -fsS "http://127.0.0.1:$PORT/health" >/dev/null 2>&1; then
  echo "port $PORT is already serving a Tablon instance; set TABLON_BENCH_PORT" >&2
  exit 1
fi
# Keep benchmark data on the repository filesystem. Some temp filesystems add
# hundreds of milliseconds to SQLite startup and make startup numbers noisy.
DATA_DIR=$(mktemp -d "$ROOT/tmp/tablon-bench.XXXXXX")
BIN=$(mktemp "$ROOT/tmp/tablond-bench.XXXXXX")
LOG="$DATA_DIR/server.log"
pid=""

cleanup() {
  if [ -n "${pid:-}" ]; then
    kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
  rm -rf "$DATA_DIR" "$BIN"
}
trap cleanup EXIT INT TERM

now() { perl -MTime::HiRes=time -e 'printf "%.6f\n", time'; }
json_field() {
  python3 -c 'import functools,json,sys; d=json.load(sys.stdin); print(functools.reduce(lambda x,k: x[k], sys.argv[1].split("."), d))' "$1"
}
rss_kb() { ps -o rss= -p "$pid" | awk '{print $1}'; }
mb() { awk -v kb="$1" 'BEGIN { printf "%.1f", kb / 1024 }'; }
elapsed() { awk -v a="$1" -v b="$2" 'BEGIN { printf "%.3f", b - a }'; }

for cmd in go curl perl python3 ps awk; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "missing required command: $cmd" >&2
    exit 1
  fi
done

if [ ! -f frontend/dist/index.html ]; then
  echo "frontend/dist is missing; run pnpm --dir frontend build first" >&2
  exit 1
fi
echo "building tablond..."
go build -trimpath -tags embed -ldflags "-s -w" -o "$BIN" ./cmd/server
# Warm the OS/binary cache once so a freshly built macOS binary does not
# include one-time code-signature/loader cost in the startup number.
"$BIN" --version >/dev/null 2>&1 || true

start=$(now)
TABLON_HOST=127.0.0.1 \
TABLON_PORT="$PORT" \
TABLON_DATA_DIR="$DATA_DIR" \
TABLON_IDLE_ENABLED=true \
TABLON_IDLE_TIMEOUT="$IDLE_TIMEOUT" \
TABLON_LOG_LEVEL=error \
  "$BIN" >"$LOG" 2>&1 &
pid=$!

ready=0
while :; do
  if curl -fsS "http://127.0.0.1:$PORT/ready" >/dev/null 2>&1; then
    ready=1
    break
  fi
  if ! kill -0 "$pid" 2>/dev/null; then
    cat "$LOG" >&2
    echo "server exited before becoming ready" >&2
    exit 1
  fi
  sleep 0.01
done
if [ "$ready" -ne 1 ]; then
  cat "$LOG" >&2
  exit 1
fi
startup=$(elapsed "$start" "$(now)")
rss_start=$(rss_kb)

project_json=$(curl -fsS -X POST "http://127.0.0.1:$PORT/api/v1/projects" \
  -H 'Content-Type: application/json' -d '{"name":"Benchmark"}')
project_id=$(printf '%s' "$project_json" | json_field id)
project_ref=$(printf '%s' "$project_json" | json_field identifier)

i=1
while [ "$i" -le 100 ]; do
  curl -fsS -X POST "http://127.0.0.1:$PORT/api/v1/projects/$project_id/issues" \
    -H 'Content-Type: application/json' \
    -d "{\"title\":\"Benchmark issue $i\"}" >/dev/null
  i=$((i + 1))
done

issue_ref=$(curl -fsS "http://127.0.0.1:$PORT/api/v1/projects/$project_id/issues?limit=1" |
  python3 -c 'import json,sys; print(json.load(sys.stdin)["items"][0]["identifier"])')

i=1
while [ "$i" -le 50 ]; do
  curl -fsS -X POST "http://127.0.0.1:$PORT/api/v1/issues/$issue_ref/comments" \
    -H 'Content-Type: application/json' \
    -d "{\"content\":\"Benchmark comment $i\"}" >/dev/null
  i=$((i + 1))
done

rss_active=$(rss_kb)

list_start=$(now)
curl -fsS "http://127.0.0.1:$PORT/api/v1/projects/$project_id/issues?limit=50" >/dev/null
list_latency=$(elapsed "$list_start" "$(now)")

sleep 5
rss_idle=$(rss_kb)
idle_state=$(curl -fsS "http://127.0.0.1:$PORT/api/v1/system/status" | json_field idle.state)

reactivate_start=$(now)
curl -fsS "http://127.0.0.1:$PORT/api/v1/projects/$project_id" >/dev/null
reactivate_latency=$(elapsed "$reactivate_start" "$(now)")
rss_reactivated=$(rss_kb)
active_state=$(curl -fsS "http://127.0.0.1:$PORT/api/v1/system/status" | json_field idle.state)

cat <<REPORT
# Tablon benchmark

- Date: $(date -u '+%Y-%m-%dT%H:%M:%SZ')
- Platform: $(uname -sm)
- Go: $(go version)
- Binary: $(wc -c <"$BIN" | awk '{printf "%.1f MB", $1 / 1048576}')
- Data: 1 project, 100 issues, 50 comments
- Idle timeout: $IDLE_TIMEOUT
- Startup is measured after a warm-up version check.

| Metric | Value |
|---|---|
| Startup to /ready | ${startup}s |
| RSS after start | $(mb "$rss_start") MB |
| RSS after workload | $(mb "$rss_active") MB |
| RSS after idle ($idle_state) | $(mb "$rss_idle") MB |
| First request after idle | ${reactivate_latency}s |
| RSS after reactivation ($active_state) | $(mb "$rss_reactivated") MB |
| List issues (50) latency | ${list_latency}s |

The server was stopped and its temporary data directory removed.
REPORT
