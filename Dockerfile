FROM alpine:3.20 AS base

ARG DOCKER_CLI_VERSION=${DOCKER_CLI_VERSION}
RUN wget -O- https://download.docker.com/linux/static/stable/$(uname -m)/docker-${DOCKER_CLI_VERSION}.tgz | \
    tar -xzf - docker/docker --strip-component=1 -C /usr/local/bin

COPY dive /usr/local/bin/

FROM scratch
COPY --from=base / /

ENTRYPOINT ["/usr/local/bin/dive"]
