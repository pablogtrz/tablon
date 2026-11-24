#!/usr/bin/env sh
# Runtime benchmark for native and Docker on Linux (Raspberry Pi class hardware).
# Measures startup, working/idle RSS, peak RSS, data storage and a small API workload.
# Requires: docker, curl, python3, ps, awk, du, stat.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
mkdir -p "$ROOT/tmp"

IMAGE=${TABLON_BENCH_IMAGE:-tablon:bench}
NATIVE_PORT=${TABLON_BENCH_NATIVE_PORT:-28790}
DOCKER_PORT=${TABLON_BENCH_DOCKER_PORT:-28791}
IDLE_TIMEOUT=${TABLON_BENCH_IDLE_TIMEOUT:-3s}
ISSUES=${TABLON_BENCH_ISSUES:-100}
COMMENTS=${TABLON_BENCH_COMMENTS:-50}
REPORT="$ROOT/tmp/bench-runtime.md"

native_dir=""
native_pid=""
docker_name=""
docker_volume=""

cleanup() {
  if [ -n "${native_pid:-}" ]; then
    kill "$native_pid" 2>/dev/null || true
    wait "$native_pid" 2>/dev/null || true
  fi
  if [ -n "${docker_name:-}" ]; then
    docker rm -f "$docker_name" >/dev/null 2>&1 || true
  fi
  if [ -n "${docker_volume:-}" ]; then
    docker volume rm "$docker_volume" >/dev/null 2>&1 || true
  fi
  if [ -n "${native_dir:-}" ]; then
    rm -rf "$native_dir"
  fi
}
trap cleanup EXIT INT TERM

now() { date +%s%N; }
elapsed() { awk -v a="$1" -v b="$2" 'BEGIN { printf "%.3f", (b - a) / 1000000000 }'; }
mib() { awk -v kb="$1" 'BEGIN { printf "%.1f", kb / 1024 }'; }
json_field() {
  python3 -c 'import functools,json,sys; d=json.load(sys.stdin); print(functools.reduce(lambda x,k: x[k], sys.argv[1].split("."), d))' "$1"
}
rss_kb() { ps -o rss= -p "$1" 2>/dev/null | awk '{print $1}'; }
stat_size() { stat -c %s "$1" 2>/dev/null || wc -c <"$1"; }
dir_kb() { du -sk "$1" | awk '{print $1}'; }

wait_ready() {
  base=$1
  pid=$2
  while :; do
    if curl -fsS "$base/ready" >/dev/null 2>&1; then
      return 0
    fi
    if [ -n "$pid" ] && ! kill -0 "$pid" 2>/dev/null; then
      return 1
    fi
    sleep 0.05
  done
}

check_port() {
  port=$1
  if curl -fsS "http://127.0.0.1:$port/health" >/dev/null 2>&1; then
    echo "port $port is already serving a Tablon instance" >&2
    exit 1
  fi
}

workload() {
  base=$1
  project=$(curl -fsS -X POST "$base/api/v1/projects" \
    -H 'Content-Type: application/json' -d '{"name":"Runtime benchmark"}')
  project_id=$(printf '%s' "$project" | json_field id)

  first=$(curl -fsS -X POST "$base/api/v1/projects/$project_id/issues" \
    -H 'Content-Type: application/json' -d '{"title":"Benchmark issue 1"}')
  issue_ref=$(printf '%s' "$first" | json_field identifier)
  w_one=$(rss_kb "$2")
  max=$w_one

  i=2
  while [ "$i" -le "$ISSUES" ]; do
    curl -fsS -X POST "$base/api/v1/projects/$project_id/issues" \
      -H 'Content-Type: application/json' \
      -d "{\"title\":\"Benchmark issue $i\"}" >/dev/null
    r=$(rss_kb "$2")
    if [ -n "$r" ] && [ "$r" -gt "$max" ]; then max=$r; fi
    i=$((i + 1))
  done

  i=1
  while [ "$i" -le "$COMMENTS" ]; do
    curl -fsS -X POST "$base/api/v1/issues/$issue_ref/comments" \
      -H 'Content-Type: application/json' \
      -d "{\"content\":\"Benchmark comment $i\"}" >/dev/null
    r=$(rss_kb "$2")
    if [ -n "$r" ] && [ "$r" -gt "$max" ]; then max=$r; fi
    i=$((i + 1))
  done

  list_start=$(now)
  curl -fsS "$base/api/v1/projects/$project_id/issues?limit=50" >/dev/null
  list_latency=$(elapsed "$list_start" "$(now)")

  w_active=$(rss_kb "$2")
  w_peak=$max
  w_list=$list_latency
}

native_bench() {
  base="http://127.0.0.1:$NATIVE_PORT"
  check_port "$NATIVE_PORT"
  native_dir=$(mktemp -d "$ROOT/tmp/tablon-native.XXXXXX")
  bin="$native_dir/tablond"
  tmp_container="tablon-bench-extract-$$"

  mkdir -p "$native_dir/data"
  docker create --name "$tmp_container" "$IMAGE" >/dev/null
  docker cp "$tmp_container:/usr/local/bin/tablond" "$bin" >/dev/null
  docker rm "$tmp_container" >/dev/null
  chmod +x "$bin"

  start=$(now)
  TABLON_HOST=127.0.0.1 \
  TABLON_PORT="$NATIVE_PORT" \
  TABLON_DATA_DIR="$native_dir/data" \
  TABLON_IDLE_ENABLED=true \
  TABLON_IDLE_TIMEOUT="$IDLE_TIMEOUT" \
  TABLON_LOG_LEVEL=error \
    "$bin" >"$native_dir/server.log" 2>&1 &
  native_pid=$!
  wait_ready "$base" "$native_pid" || {
    cat "$native_dir/server.log" >&2
    echo "native server did not become ready" >&2
    exit 1
  }
  native_startup=$(elapsed "$start" "$(now)")
  native_start=$(rss_kb "$native_pid")

  workload "$base" "$native_pid" "$native_dir"
  native_one=$w_one
  native_active=$w_active
  native_peak=$w_peak
  native_list=$w_list
  native_data_active=$(dir_kb "$native_dir/data")

  sleep 5
  native_idle=$(rss_kb "$native_pid")
  native_state=$(curl -fsS "$base/api/v1/system/status" | json_field idle.state)
  native_data_idle=$(dir_kb "$native_dir/data")

  react_start=$(now)
  curl -fsS "$base/api/v1/projects" >/dev/null
  native_react_latency=$(elapsed "$react_start" "$(now)")
  native_react=$(rss_kb "$native_pid")

  native_binary_bytes=$(stat_size "$bin")

  kill "$native_pid" 2>/dev/null || true
  wait "$native_pid" 2>/dev/null || true
  native_pid=""
  native_data_stopped=$(dir_kb "$native_dir/data")
}

docker_bench() {
  base="http://127.0.0.1:$DOCKER_PORT"
  check_port "$DOCKER_PORT"
  docker_volume="tablon-bench-data-$$"
  docker_name="tablon-bench-$$"
  docker_out="$ROOT/tmp/tablon-docker-$$"
  mkdir -p "$docker_out"

  docker volume create "$docker_volume" >/dev/null
  start=$(now)
  docker run -d --name "$docker_name" \
    --read-only --tmpfs /tmp \
    --cap-drop ALL --security-opt no-new-privileges:true \
    -p "127.0.0.1:$DOCKER_PORT:8448" \
    -v "$docker_volume:/data" \
    -e TABLON_IDLE_ENABLED=true \
    -e TABLON_IDLE_TIMEOUT="$IDLE_TIMEOUT" \
    -e TABLON_LOG_LEVEL=error \
    "$IMAGE" >/dev/null
  docker_pid=$(docker inspect -f '{{.State.Pid}}' "$docker_name")
  wait_ready "$base" "$docker_pid" || {
    docker logs "$docker_name" >&2 || true
    echo "docker server did not become ready" >&2
    exit 1
  }
  docker_startup=$(elapsed "$start" "$(now)")
  docker_start=$(rss_kb "$docker_pid")

  workload "$base" "$docker_pid" "$docker_out"
  docker_one=$w_one
  docker_active=$w_active
  docker_peak=$w_peak
  docker_list=$w_list
  docker_volume_active=$(docker exec "$docker_name" du -sk /data | awk '{print $1}')

  sleep 5
  docker_idle=$(rss_kb "$docker_pid")
  docker_state=$(curl -fsS "$base/api/v1/system/status" | json_field idle.state)
  docker_volume_idle=$(docker exec "$docker_name" du -sk /data | awk '{print $1}')

  react_start=$(now)
  curl -fsS "$base/api/v1/projects" >/dev/null
  docker_react_latency=$(elapsed "$react_start" "$(now)")
  docker_react=$(rss_kb "$docker_pid")

  docker_image_bytes=$(docker image inspect --format '{{.Size}}' "$IMAGE")
  docker_digest=$(docker image inspect --format '{{index .RepoDigests 0}}' "$IMAGE" 2>/dev/null || docker image inspect --format '{{.Id}}' "$IMAGE")

  docker rm -f "$docker_name" >/dev/null
  docker_name=""
  docker volume rm "$docker_volume" >/dev/null
  docker_volume=""
  rm -rf "$docker_out"
}

for cmd in docker curl python3 ps awk du stat; do
  command -v "$cmd" >/dev/null 2>&1 || {
    echo "missing required command: $cmd" >&2
    exit 1
  }
done

docker image inspect "$IMAGE" >/dev/null 2>&1 || {
  echo "missing image $IMAGE; build it first (docker build -t $IMAGE .)" >&2
  exit 1
}

source_kb=$(du -sk --exclude=.git --exclude=tmp --exclude=frontend/dist . | awk '{print $1}')

native_bench
docker_bench

cat >"$REPORT" <<REPORT
# Tablon runtime benchmark (native vs Docker)

- Date: $(date -u '+%Y-%m-%dT%H:%M:%SZ')
- Platform: $(uname -sm)
- Kernel: $(uname -r)
- Docker: $(docker --version)
- Image: \`$IMAGE\` (${docker_digest})
- Idle timeout: \`$IDLE_TIMEOUT\`
- Workload: 1 project, $ISSUES issues, $COMMENTS comments
- Source checkout (excluding \`.git\`, \`tmp/\`, \`frontend/dist/\`): $(mib "$source_kb") MiB

| Metric | Native | Docker |
|---|---:|---:|
| Startup to \`/ready\` (s) | ${native_startup} | ${docker_startup} |
| Binary / image size (MiB) | $(mib "$((native_binary_bytes / 1024))") | $(mib "$((docker_image_bytes / 1024))") |
| RSS after start (MiB) | $(mib "$native_start") | $(mib "$docker_start") |
| RSS after first issue (MiB) | $(mib "$native_one") | $(mib "$docker_one") |
| Peak RSS during workload (MiB) | $(mib "$native_peak") | $(mib "$docker_peak") |
| RSS after workload (MiB) | $(mib "$native_active") | $(mib "$docker_active") |
| RSS after idle, state \`$native_state\` (MiB) | $(mib "$native_idle") | $(mib "$docker_idle") |
| First request after idle (s) | ${native_react_latency} | ${docker_react_latency} |
| RSS after reactivation, state \`active\` (MiB) | $(mib "$native_react") | $(mib "$docker_react") |
| List 50 issues latency (s) | ${native_list} | ${docker_list} |
| Data / volume after workload (MiB) | $(mib "$native_data_active") | $(mib "$docker_volume_active") |
| Data / volume after idle (MiB) | $(mib "$native_data_idle") | $(mib "$docker_volume_idle") |
| Data / volume after shutdown (MiB) | $(mib "$native_data_stopped") | n/a |

Native data directory: \`$(basename "$native_dir")\`. Docker volume was removed after the measurement.
REPORT

cat "$REPORT"
