FROM golang:latest

ADD . /go/src/github.com/iamthebot/jumphasher

RUN go install github.com/iamthebot/jumphasher/api

VOLUME /mnt/ssl

ENTRYPOINT ["/go/bin/api","-sslcert","/mnt/ssl/server.crt","-sslkey","/mnt/ssl/server.pem"]

EXPOSE 443
EXPOSE 80