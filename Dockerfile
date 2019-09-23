FROM alpine:3.10
COPY dist/dive_linux_amd64/dive /
ENTRYPOINT ["/dive"]
