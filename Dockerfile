FROM debian:stable-slim
COPY dive /
ENTRYPOINT ["/dive"]
