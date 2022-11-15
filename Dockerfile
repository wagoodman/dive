FROM alpine:3.16.3@sha256:b95359c2505145f16c6aa384f9cc74eeff78eb36d308ca4fd902eeeb0a0b161b

ARG DOCKER_CLI_VERSION=${DOCKER_CLI_VERSION}
RUN wget -O- https://download.docker.com/linux/static/stable/$(uname -m)/docker-${DOCKER_CLI_VERSION}.tgz | \
    tar -xzf - docker/docker --strip-component=1 && \
    mv docker /usr/local/bin

COPY dive /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/dive"]
