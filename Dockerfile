FROM golang:1.25-alpine@sha256:26111811bc967321e7b6f852e914d14bede324cd1accb7f81811929a6a57fea9 AS builder

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


FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

LABEL org.opencontainers.image.authors="Knut Ahlers <knut@ahlers.me>" \
      org.opencontainers.image.url="https://github.com/users/Luzifer/packages/container/package/twitch-bot" \
      org.opencontainers.image.documentation="https://luzifer.github.io/twitch-bot/" \
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
