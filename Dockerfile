FROM alpine:3.17.2@sha256:8a81a7c2af9caf25aa960c6ef70e198b2d7194841546d3b26c9a1eb308a360b3

ARG DOCKER_CLI_VERSION=${DOCKER_CLI_VERSION}
RUN wget -O- https://download.docker.com/linux/static/stable/$(uname -m)/docker-${DOCKER_CLI_VERSION}.tgz | \
    tar -xzf - docker/docker --strip-component=1 && \
    mv docker /usr/local/bin

COPY dive /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/dive"]
