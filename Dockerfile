FROM alpine:latest
ADD README.md /somefile.txt
RUN cp /somefile.txt /root/somefile.txt