FROM golang:1.26.4-alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.9.0@sha256:7f6aa7706c898b28e43265e859bdd62d0855fcb8f0777d19e6b0e33963766394 . /

COPY . /go/src/twitch-bot
WORKDIR /go/src/twitch-bot

ENV CGO_ENABLED=0 \
    GOPATH=/go

RUN set -ex \
 && apk --no-cache add \
      curl \
      git \
      make \
      nodejs \
      npm \
 && git config --global --add safe.directory /go/src/twitch-bot \
 && make node_modules frontend_prod \
 && go install \
      -trimpath \
      -mod=readonly \
      -modcacherw \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)"


FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

LABEL org.opencontainers.image.authors="Knut Ahlers <knut@ahlers.me>" \
      org.opencontainers.image.url="https://github.com/users/Luzifer/packages/container/package/twitch-bot" \
      org.opencontainers.image.documentation="https://twitch-bot-docs.luzifer.io/" \
      org.opencontainers.image.source="https://github.com/Luzifer/twitch-bot" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.title="Self-hosted alternative to one of the big Twitch bots managed by big companies"

ENV CONFIG=/data/config.yaml \
    STORAGE_CONN_STRING=/data/store.db

RUN set -ex \
 && apk --no-cache add \
      bash \
      ca-certificates \
      curl \
      jq \
      tzdata \
 && mkdir /data \
 && chown 1000:1000 /data

COPY --from=builder /go/bin/twitch-bot /usr/local/bin/twitch-bot

USER 1000:1000
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/twitch-bot"]
CMD ["--"]

# vim: set ft=Dockerfile:
