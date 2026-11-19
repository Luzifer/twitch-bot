FROM golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.28.12@sha256:c320a949a13747902e9d71296e886d810ddb733b57d0a341db8a11e246c4b43c . /

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


FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

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
