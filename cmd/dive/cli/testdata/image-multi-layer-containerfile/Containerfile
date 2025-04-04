FROM busybox:1.37.0@sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f
ADD example.md /somefile.txt
RUN mkdir -p /root/example/really/nested
RUN cp /somefile.txt /root/example/somefile1.txt
RUN chmod 444 /root/example/somefile1.txt
RUN cp /somefile.txt /root/example/somefile2.txt
RUN cp /somefile.txt /root/example/somefile3.txt
RUN mv /root/example/somefile3.txt /root/saved.txt
RUN cp /root/saved.txt /root/.saved.txt
RUN chmod +x /root/saved.txt
RUN chmod 421 /root
RUN rm -rf /root/example/
ADD overwrite.md /root/saved.txt
