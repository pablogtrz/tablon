# Self-hosting Tablon

Tablon runs as one Go binary with the SPA embedded, or as a Docker Compose
service. It needs no external database or Node runtime. SQLite lives in
`TABLON_DATA_DIR`.

## Docker Compose

```bash
cp .env.example .env
# Edit .env and set TABLON_API_TOKEN before exposing Tablon beyond loopback.
docker compose up -d --build
docker compose ps
docker compose logs -f
```

Open `http://127.0.0.1:8448`. Compose publishes only on `127.0.0.1` by default
and stores data in the `tablon-data` named volume. `docker compose down` keeps
the volume; `docker compose down -v` deletes it.

- `TABLON_HOST_PORT` changes the host port; the container listens on `8448`.
- To expose on a LAN, change the Compose binding from `127.0.0.1` to `0.0.0.0`
  and set `TABLON_API_TOKEN`; prefer a TLS reverse proxy.
- The runtime image is Alpine, runs as UID/GID 10001, uses a read-only root and
  keeps `/data` and `/tmp` writable.

## Prebuilt binary

Download the bundle archive from the release page (it contains `tablond` and
`tablon`):

```bash
# Replace os/arch with your platform, for example linux_amd64 or darwin_arm64.
curl -LO https://github.com/pablogtrz/tablon/releases/latest/download/tablon_VERSION_OS_ARCH.tar.gz
curl -LO https://github.com/pablogtrz/tablon/releases/latest/download/SHA256SUMS
shasum -a 256 -c SHA256SUMS     # macOS
sha256sum -c SHA256SUMS          # Linux
tar -xzf tablon_VERSION_OS_ARCH.tar.gz
TABLON_DATA_DIR=./data ./tablond
```

Open `http://127.0.0.1:8448`. The CLI is published separately as
`tablon-cli_VERSION_OS_ARCH.tar.gz`; alternatively run
`go install github.com/pablogtrz/tablon/cmd/tablon@latest`.

Verify a binary without starting the server:

```bash
./tablond --version
./tablon version
```

## Configuration

| Variable | Default | Notes |
|---|---|---|
| `TABLON_HOST` | `127.0.0.1` | Bind address. Use `0.0.0.0` only behind a trusted proxy or with a token. |
| `TABLON_PORT` | `8448` | API/UI port. |
| `TABLON_DATA_DIR` | `./data` | SQLite directory. Must be writable by the process user. |
| `TABLON_API_TOKEN` | empty | Optional bearer token for `/api/v1`. Empty means no auth. |
| `TABLON_IDLE_ENABLED` | `true` | Enable the idle activity gate. |
| `TABLON_IDLE_TIMEOUT` | `5m` | Quiet period before idle. Parsed as a Go duration (`300s`, `5m`). |
| `TABLON_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. |
| `TABLON_HOST_PORT` | `8448` | Compose-only host port. |
| `TABLON_CLI_URL` | `http://127.0.0.1:8448` | CLI server URL. |
| `TABLON_CLI_TOKEN` | empty | CLI bearer token. |

Invalid core configuration fails at startup with a clear error. The API contract
is available at `/openapi.yaml`; see [`docs/api/openapi.yaml`](api/openapi.yaml).

## Security

Tablon has no accounts, roles or multi-user authorization. The optional bearer
token grants full API access.

- **Keep it on loopback** unless you need remote access. Compose does this by
  default.
- **Set `TABLON_API_TOKEN` before exposing it.** Without a token, anyone who can
  reach the port can read and modify all projects.
- **Use HTTPS through a reverse proxy.** Tablon does not terminate TLS itself.
- **Rotate the token** by restarting the process with a new value. Existing
  browser sessions must re-enter it.
- **The UI stores the token in `sessionStorage`.** Closing the browser session
  clears it. There is no server-side session store.
- Public paths: the SPA assets, `/health`, `/ready` and `/openapi.yaml`. All
  `/api/v1` reads and writes are gated when a token is set.
- Logs never include the token. Do not put secrets in `.env.example` or commit
  `.env`.
- The image runs as UID/GID 10001. Compose additionally uses a read-only root,
  `cap_drop: ALL` and `no-new-privileges`. Do not expose the data directory over
  the network.
- Security headers and a same-origin CSP are set by the Go server. There is no
  cross-origin browser API; agents should use the CLI or server-to-server REST.

### Reverse proxy

Caddy:

```caddyfile
tablon.example.com {
    reverse_proxy 127.0.0.1:8448 {
        flush_interval -1
    }
}
```

nginx:

```nginx
server {
    listen 443 ssl http2;
    server_name tablon.example.com;

    location / {
        proxy_pass http://127.0.0.1:8448;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE: do not buffer, allow long-lived streams.
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 1h;
    }
}
```

Keep `Authorization` intact, do not cache `/api/`, and do not strip the
`X-Accel-Buffering: no` header set by the SSE endpoint. A subpath deployment is
not supported in v0.1; serve Tablon at the root of a host.

### systemd (binary)

```ini
[Unit]
Description=Tablon
After=network.target

[Service]
User=tablon
Group=tablon
EnvironmentFile=/etc/tablon.env
WorkingDirectory=/var/lib/tablon
ExecStart=/usr/local/bin/tablond
Restart=on-failure
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/var/lib/tablon

[Install]
WantedBy=multi-user.target
```

`/etc/tablon.env` should contain `TABLON_DATA_DIR=/var/lib/tablon`,
`TABLON_HOST=127.0.0.1` and, if needed, `TABLON_API_TOKEN`. Put a TLS reverse
proxy in front for remote access.

## Health and idle

- `GET /health` is a public liveness check and does not touch SQLite. The
  container healthcheck uses it.
- `GET /ready` is a public database readiness check. It may open a SQLite
  connection; use it for orchestrator readiness, not as a high-frequency idle
  probe.
- `/api/v1/system/status` reports version, auth, database and idle state. It does
  not reset idle mode; while idle it reports `database: "idle"` without pinging.
- Idle mode keeps HTTP and SQLite available. It releases the idle writer
  connection and runs `PRAGMA shrink_memory`; it does not stop the process.
- An open SSE stream keeps the server active. Reconnect/refetch after network
  loss; v0.1 has no event replay.

## Backups

SQLite runs in WAL mode. Prefer a consistent backup while the server is running:

```bash
sqlite3 /path/to/data/app.db ".backup '/backup/tablon-$(date +%F).db'"
sqlite3 /backup/tablon-$(date +%F).db "PRAGMA integrity_check;"
```

For a Docker named volume without `sqlite3` on the host, stop the container and
copy the volume:

```bash
docker compose stop
docker run --rm -v tablon-data:/data -v "$PWD":/backup alpine \
  tar czf /backup/tablon-data.tgz -C /data .
docker compose start
```

Do not copy a live `app.db` without its WAL/SHM files. To restore, stop Tablon,
replace the data directory contents, and ensure the UID that runs Tablon (10001
in the container) can write the files. Verify with `/ready` and the UI.

## Upgrades and rollback

1. Back up the database first.
2. Pull a tagged release (`TABLON_VERSION=0.1.0 docker compose up -d`) or replace
   the binary.
3. Start Tablon. Embedded migrations run automatically at startup.
4. Verify `/ready`, the Settings page and a write operation.

Migrations are forward-only. A rollback means restoring the pre-upgrade backup
and running the previous version. Pin a semantic version in production instead
of `latest`.

## Troubleshooting

| Symptom | Check |
|---|---|
| `401 unauthorized` | Set the token in Settings or pass `Authorization: Bearer ...`. |
| Port already in use | Change `TABLON_PORT` (binary) or `TABLON_HOST_PORT` (Compose). |
| `permission denied` on `/data` | The container user is UID 10001; a bind mount must be writable by it. |
| Container unhealthy | Check `docker compose logs`; `/health` is liveness, `/ready` is the DB check. |
| SSE not live | Disable proxy buffering and allow long read timeouts. |
| Status says `idle` | Normal after `TABLON_IDLE_TIMEOUT`; the next API request reactivates. |

## Release notes

See [`docs/release.md`](release.md) for the tag, artifact and Docker image
runbook.
