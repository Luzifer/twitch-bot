FROM golang:1.27.3-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.28.6@sha256:d35ec722ba2d0bc7d8ff9d9920d2b9f851e566d13822e33da1272d085a5b03a8 . /

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
      org.opencontainers.image.version='3.44.3'

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
