#!/usr/bin/env sh

set -e

cd "$(dirname $0)"

docker build -t wagoodman/dive:dev - <<EOF
FROM golang:alpine AS build
RUN apk add -U --no-cache gpgme-dev gcc musl-dev btrfs-progs-dev lvm2-dev curl git \
 && curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh \
 && curl -L https://download.docker.com/linux/static/stable/x86_64/docker-19.03.1.tgz | tar -xzf - docker/docker --strip-component=1 -C /usr/local/bin
EOF

docker run --rm \
  -v //var/run/docker.sock://var/run/docker.sock \
  -v /$(pwd)://src \
  -w //src \
  wagoodman/dive:dev \
  goreleaser \
    -f .goreleaser.docker.yml \
    --snapshot \
    --skip-publish \
    --rm-dist

sudo chown -R $USER:$USER dist
