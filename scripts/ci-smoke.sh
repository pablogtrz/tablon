#!/usr/bin/env sh
# Build the single binary with the embedded SPA and smoke-test it without Node.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
mkdir -p "$ROOT/tmp"

PORT=${TABLON_SMOKE_PORT:-18081}
TOKEN=smoke-token
if curl -fsS "http://127.0.0.1:$PORT/health" >/dev/null 2>&1; then
  echo "port $PORT is already serving a Tablon instance; set TABLON_SMOKE_PORT" >&2
  exit 1
fi
DATA_DIR=$(mktemp -d "$ROOT/tmp/ci-smoke.XXXXXX")
BIN="$ROOT/tmp/tablond-ci"
pid=""

cleanup() {
  if [ -n "${pid:-}" ]; then
    kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
  rm -rf "$DATA_DIR" "$BIN"
}
trap cleanup EXIT INT TERM

if [ ! -f frontend/dist/index.html ]; then
  echo "frontend/dist is missing; run pnpm --dir frontend build first" >&2
  exit 1
fi

VERSION=${VERSION:-ci}
COMMIT=${GITHUB_SHA:-unknown}
DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
go build -trimpath -tags embed \
  -ldflags "-X github.com/pablogtrz/tablon/internal/version.Version=${VERSION} \
    -X github.com/pablogtrz/tablon/internal/version.Commit=${COMMIT} \
    -X github.com/pablogtrz/tablon/internal/version.Date=${DATE}" \
  -o "$BIN" ./cmd/server

TABLON_HOST=127.0.0.1 \
TABLON_PORT="$PORT" \
TABLON_DATA_DIR="$DATA_DIR" \
TABLON_API_TOKEN="$TOKEN" \
TABLON_IDLE_ENABLED=false \
TABLON_LOG_LEVEL=error \
  "$BIN" >"$DATA_DIR/server.log" 2>&1 &
pid=$!

ready=0
while :; do
  if curl -fsS "http://127.0.0.1:$PORT/ready" >/dev/null 2>&1; then
    ready=1
    break
  fi
  if ! kill -0 "$pid" 2>/dev/null; then
    cat "$DATA_DIR/server.log" >&2
    exit 1
  fi
  sleep 0.02
done
if [ "$ready" -ne 1 ]; then
  cat "$DATA_DIR/server.log" >&2
  exit 1
fi

assert_status() {
  path=$1
  want=$2
  shift 2
  got=$(curl -s -o /dev/null -w '%{http_code}' "$@" "http://127.0.0.1:$PORT$path")
  if [ "$got" != "$want" ]; then
    echo "GET $path = $got, want $want" >&2
    exit 1
  fi
}

curl -fsS "http://127.0.0.1:$PORT/" | grep -qi '<!doctype html>'
curl -fsS "http://127.0.0.1:$PORT/projects/TABLON/issues/TABLON-1" | grep -qi '<!doctype html>'
curl -fsS "http://127.0.0.1:$PORT/favicon.svg" >/dev/null
curl -fsS "http://127.0.0.1:$PORT/openapi.yaml" | grep -q 'openapi: 3.0.3'
curl -fsS -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:$PORT/api/v1/system/status" | grep -q '"idle"'
assert_status "/api/v1/not-a-route" 404 -H "Authorization: Bearer $TOKEN"
assert_status "/assets/not-a-file.js" 404

body=$(curl -s -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:$PORT/api/v1/not-a-route")
printf '%s' "$body" | grep -q '"code":"not_found"'

echo "binary smoke ok"
