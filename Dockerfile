FROM golang:latest
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /go/src/github.com/b3log/wide
WORKDIR /go/src/github.com/b3log/wide

CMD ["./docker.sh"]

ENV GOROOT /usr/src/go

EXPOSE 7070