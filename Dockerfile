FROM golang:1.26.4-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS builder

COPY --from=ghcr.io/luzifer-docker/pnpm:v11.5.3@sha256:d2c5a4b46d7f214c92342ebaa9ae1439faf9315e0334a10f77b3231602e42e39 . /

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


FROM alpine:3.24@sha256:f5064d3e5f88c467c714509f491853ab2d951932c5cad699c0cb969dcec6f3b4

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
