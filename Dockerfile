FROM golang:alpine as builder

COPY . /go/src/github.com/Luzifer/twitch-bot
WORKDIR /go/src/github.com/Luzifer/twitch-bot

RUN set -ex \
 && apk add --update git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly

FROM alpine:latest

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

ENV CONFIG=/data/config.yaml \
    STORAGE_FILE=/data/store.json.gz

RUN set -ex \
 && apk --no-cache add \
      bash \
      ca-certificates \
      curl \
      jq

COPY --from=builder /go/bin/twitch-bot /usr/local/bin/twitch-bot

VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/twitch-bot"]
CMD ["--"]

# vim: set ft=Dockerfile:
