FROM debian:stretch

RUN apt update && \
     apt install -y \
     curl \
     && rm -rf /var/lib/apt/lists/*

RUN curl -L -O https://github.com/wagoodman/dive/releases/download/v0.0.5/dive_0.0.5_linux_amd64.deb
RUN apt install ./dive_0.0.5_linux_amd64.deb

ENTRYPOINT ["dive"]
