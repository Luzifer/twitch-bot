FROM luzifer/archlinux as builder

COPY . /go/src/github.com/Luzifer/twitch-bot
WORKDIR /go/src/github.com/Luzifer/twitch-bot

ENV CGO_ENABLED=0 \
    GOPATH=/go \
    NODE_ENV=production

RUN set -ex \
 && pacman -Syy --noconfirm \
      curl \
      git \
      go \
      make \
      nodejs-lts-fermium \
      npm \
 && make frontend \
 && go install \
      -trimpath \
      -buildmode=pie \
      -mod=readonly \
      -modcacherw \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)"


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

COPY --from=builder /go/bin/twitch-bot /usr/local/bin/twitch-bot

VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/twitch-bot"]
CMD ["--"]

# vim: set ft=Dockerfile:
