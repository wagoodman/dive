FROM alpine:3.9
COPY dive /
ENTRYPOINT ["/dive"]
