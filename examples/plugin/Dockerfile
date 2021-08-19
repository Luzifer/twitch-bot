FROM golang:alpine as builder

COPY . /go/src/github.com/Luzifer/twitch-bot/examples/plugin
WORKDIR /go/src/github.com/Luzifer/twitch-bot/examples/plugin

RUN set -ex \
 && apk add --update \
      build-base \
      git \
 && go install \
      -mod=readonly \
      github.com/Luzifer/twitch-bot@$(grep 'twitch-bot ' go.mod | sed -E 's/.*-([^-]+)$/\1/') \
 && go build \
      -buildmode=plugin \
      -mod=readonly \
      -o /go/bin/example-plugin.so


FROM alpine:latest

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

ENV CONFIG=/data/config.yaml \
    STORAGE_FILE=/data/store.json.gz

RUN set -ex \
 && apk --no-cache add \
      bash \
      ca-certificates \
      curl \
      jq \
      tzdata

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/example-plugin.so /usr/lib/twitch-bot/example-plugin.so
COPY --from=builder /go/bin/twitch-bot        /usr/local/bin/twitch-bot
COPY                config.yaml               /data/config.yaml

ENTRYPOINT ["/usr/local/bin/twitch-bot"]
CMD ["--"]

# vim: set ft=Dockerfile:
