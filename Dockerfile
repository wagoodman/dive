FROM debian:sid-slim
RUN apt-get update && apt-get install -y \
        curl \
        libdevmapper1.02.1 \
        libgpgme11-dev \
     && rm -rf /var/lib/apt/lists/* \
     && ln -s /lib/x86_64-linux-gnu/libdevmapper.so.1.02.1 /usr/lib/libdevmapper.so.1.02
ARG DOCKER_CLI_VERSION="19.03.1"
RUN curl -L https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_CLI_VERSION.tgz | \
    tar -xz  --strip-component=1 -C /usr/local/bin/ docker/docker
COPY dive /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/dive"]
