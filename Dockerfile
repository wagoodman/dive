FROM alpine:3.10

ARG DOCKER_CLI_VERSION=19.03.1
RUN wget -O- https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_CLI_VERSION}.tgz | \
    tar -xzf - docker/docker --strip-component=1 && \
    mv docker /usr/local/bin

COPY dive /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/dive"]
