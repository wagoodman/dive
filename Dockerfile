FROM alpine:3.21 AS base

ARG DOCKER_CLI_VERSION=${DOCKER_CLI_VERSION}
RUN wget -O- https://download.docker.com/linux/static/stable/$(uname -m)/docker-${DOCKER_CLI_VERSION}.tgz | \
    tar -xzf - docker/docker --strip-component=1 -C /usr/local/bin

COPY dive /usr/local/bin/

# though we could make this a multi-stage image and copy the binary to scratch, this image is small enough
# and users are expecting to be able to exec into it
ENTRYPOINT ["/usr/local/bin/dive"]
