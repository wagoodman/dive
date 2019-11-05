FROM alpine:3.10
COPY --from=wagoodman/dive:dev /usr/local/bin/docker /usr/local/bin/
COPY dive /usr/local/bin/
RUN apk add -U --no-cache gpgme device-mapper
ENTRYPOINT ["/usr/local/bin/dive"]
