FROM debian:stable-slim

RUN apt update && \
     apt install -y \
     wget \
     && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/wagoodman/dive/releases/download/v0.0.5/dive_0.0.5_linux_amd64.deb
RUN apt install ./dive_0.0.5_linux_amd64.deb

ENTRYPOINT ["dive"]
