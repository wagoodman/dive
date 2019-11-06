#!/usr/bin/env sh

set -e

cd "$(dirname $0)"

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
