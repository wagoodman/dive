FROM busybox:latest
ADD README.md /somefile.txt
RUN mkdir -p /root/example/really/nested
RUN cp /somefile.txt /root/example/somefile1.txt
RUN chmod 444 /root/example/somefile1.txt
RUN cp /somefile.txt /root/example/somefile2.txt
RUN cp /somefile.txt /root/example/somefile3.txt
RUN mv /root/example/somefile3.txt /root/saved.txt
RUN cp /root/saved.txt /root/.saved.txt
RUN rm -rf /root/example/
ADD .scripts/ /root/.data/
RUN cp /root/saved.txt /tmp/saved.again1.txt
RUN cp /root/saved.txt /root/.data/saved.again2.txt
RUN chmod +x /root/saved.txt
