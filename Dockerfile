FROM alpine:3.10
COPY dive /
ENTRYPOINT ["/dive"]
