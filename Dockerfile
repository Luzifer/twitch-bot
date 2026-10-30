FROM golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.28.8@sha256:8dbed3153972eae99ea1914a5972f70ea3d19fbd115730673239a135f7cfa61a . /

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


FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

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
