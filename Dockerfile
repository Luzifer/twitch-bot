FROM golang:1.27.1-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v12.3.4@sha256:8359cbcf2c16ca7dcafd217e9afaeae66a0fc752099c868221cc22d2587ea60b . /

COPY . /go/src/twitch-bot
WORKDIR /go/src/twitch-bot

ENV CGO_ENABLED=0 \
    GOPATH=/go

RUN <<-EOF
  set -ex

  apk --no-cache add \
    curl \
    git \
    make \
    nodejs \
    npm

  git config --global --add safe.directory /go/src/twitch-bot

  make build_prod

  install -Dm0755 -t /rootfs/usr/local/bin twitch-bot
EOF


FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

LABEL org.opencontainers.image.authors="Knut Ahlers <knut@ahlers.me>" \
      org.opencontainers.image.url="https://github.com/users/Luzifer/packages/container/package/twitch-bot" \
      org.opencontainers.image.documentation="https://twitch-bot-docs.luzifer.io/" \
      org.opencontainers.image.source="https://github.com/Luzifer/twitch-bot" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.title="Self-hosted alternative to one of the big Twitch bots managed by big companies" \
      org.opencontainers.image.version='3.44.0'

ENV CONFIG=/data/config.yaml \
    STORAGE_CONN_STRING=/data/store.db

RUN <<-EOF
  set -ex

  apk --no-cache add \
    bash \
    ca-certificates \
    curl \
    jq \
    tzdata

  mkdir /data
  chown 1000:1000 /data
EOF

COPY --from=builder /rootfs/ /

USER 1000:1000
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/twitch-bot"]
CMD ["--"]

# vim: set ft=Dockerfile:
