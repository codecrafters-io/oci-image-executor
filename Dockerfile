FROM alpine
ENV TEST=yes
CMD /bin/ash -c "echo hey"

RUN apk add haveged