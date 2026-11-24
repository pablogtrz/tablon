# syntax=docker/dockerfile:1

FROM node:22-alpine AS frontend
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@11.2.2 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

FROM golang:1.27.4-alpine AS build
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/frontend/dist ./frontend/dist
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -tags embed \
    -ldflags "-s -w \
      -X github.com/pablogtrz/tablon/internal/version.Version=${VERSION} \
      -X github.com/pablogtrz/tablon/internal/version.Commit=${COMMIT} \
      -X github.com/pablogtrz/tablon/internal/version.Date=${DATE}" \
    -o /out/tablond ./cmd/server

FROM alpine:3.22
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN apk add --no-cache ca-certificates \
    && addgroup -g 10001 -S tablon \
    && adduser -u 10001 -S -G tablon tablon \
    && mkdir -p /data /usr/local/share/doc/tablon \
    && chown -R tablon:tablon /data
COPY --from=build /out/tablond /usr/local/bin/tablond
COPY LICENSE README.md /usr/local/share/doc/tablon/
WORKDIR /data
USER 10001:10001
ENV TABLON_HOST=0.0.0.0 \
    TABLON_PORT=8080 \
    TABLON_DATA_DIR=/data \
    TMPDIR=/tmp \
    HOME=/tmp
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1
LABEL org.opencontainers.image.title="Tablon" \
      org.opencontainers.image.description="Lightweight, AI-first issue tracker" \
      org.opencontainers.image.source="https://github.com/pablogtrz/tablon" \
      org.opencontainers.image.licenses="AGPL-3.0-only" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${DATE}"
STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/tablond"]
