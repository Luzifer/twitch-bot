FROM debian:13.5@sha256:4ae67669760b807c19f23902a3fd7c121a6a70cf2ae709035674b23e712e4d62 AS builder

# renovate: pkgName=Luzifer/gziphttp datasource=github-releases
ENV GZIPHTTP_VERSION=0.1.4

SHELL ["/usr/bin/env", "bash", "-euxo", "pipefail", "-c"]

RUN <<-EOF
  apt-get update
  apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    make
EOF

COPY ./ /src/
WORKDIR /src

RUN <<-EOF
  make render_docs

  install -dm0755 /rootfs/usr/share
  cp -r ./.dist /rootfs/usr/share/twitch-bot-docs
EOF

RUN <<-EOF
  install -dm0755 /rootfs/usr/bin
  curl -sSfL "https://github.com/Luzifer/gziphttp/releases/download/v${GZIPHTTP_VERSION}/gziphttp_linux_amd64.tgz" |
    tar -xz -C /rootfs/usr/bin
EOF


FROM scratch

COPY --from=builder /rootfs/ /

ENTRYPOINT ["/usr/bin/gziphttp"]
CMD ["--serve-dir", "/usr/share/twitch-bot-docs"]
