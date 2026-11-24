#!/usr/bin/env sh
# Smoke-test a built Tablon image: non-root, no Node, persistence, health and API.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

IMAGE=${1:-tablon:ci}
PORT=${2:-18082}
NAME="tablon-ci-$$"
VOLUME="tablon-ci-data-$$"
TOKEN=ci-token
BASE="http://127.0.0.1:$PORT"

cleanup() {
  docker rm -f "$NAME" >/dev/null 2>&1 || true
  docker volume rm "$VOLUME" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

json_field() {
  python3 -c 'import functools,json,sys; d=json.load(sys.stdin); print(functools.reduce(lambda x,k: x[k], sys.argv[1].split("."), d))' "$1"
}

if curl -fsS "$BASE/health" >/dev/null 2>&1; then
  echo "port $PORT is already serving a Tablon instance" >&2
  exit 1
fi

id_output=$(docker run --rm --entrypoint id "$IMAGE")
printf '%s' "$id_output" | grep -q 'uid=10001'
if docker run --rm --entrypoint sh "$IMAGE" -c 'command -v node' | grep -q .; then
  echo "node must not be present in the runtime image" >&2
  exit 1
fi

docker volume create "$VOLUME" >/dev/null
docker run -d --name "$NAME" \
  --read-only --tmpfs /tmp \
  --cap-drop ALL --security-opt no-new-privileges:true \
  -p "127.0.0.1:$PORT:8080" \
  -v "$VOLUME:/data" \
  -e TABLON_API_TOKEN="$TOKEN" \
  -e TABLON_IDLE_ENABLED=false \
  "$IMAGE" >/dev/null

healthy=0
for _ in $(seq 1 60); do
  status=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$NAME" 2>/dev/null || echo missing)
  if [ "$status" = healthy ]; then
    healthy=1
    break
  fi
  sleep 1
done
if [ "$healthy" -ne 1 ]; then
  docker logs "$NAME" >&2 || true
  echo "container did not become healthy" >&2
  exit 1
fi

curl -fsS "$BASE/" | grep -qi '<!doctype html>'
curl -fsS "$BASE/health" >/dev/null
curl -fsS "$BASE/ready" >/dev/null
curl -fsS "$BASE/openapi.yaml" | grep -q 'openapi: 3.0.3'
curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/system/status" | grep -q '"idle"'

project=$(curl -fsS -X POST "$BASE/api/v1/projects" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Docker smoke"}')
project_id=$(printf '%s' "$project" | json_field id)
project_ref=$(printf '%s' "$project" | json_field identifier)

docker restart "$NAME" >/dev/null
healthy=0
for _ in $(seq 1 60); do
  status=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$NAME" 2>/dev/null || echo missing)
  if [ "$status" = healthy ]; then
    healthy=1
    break
  fi
  sleep 1
done
if [ "$healthy" -ne 1 ]; then
  docker logs "$NAME" >&2 || true
  echo "container did not become healthy after restart" >&2
  exit 1
fi

curl -fsS -H "Authorization: Bearer $TOKEN" "$BASE/api/v1/projects/$project_id" | grep -q "$project_ref"
echo "docker smoke ok"
