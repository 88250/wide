FROM golang:latest
MAINTAINER Liang Ding <dl88250@gmail.com>

RUN go get github.com/88250/ide_stub
RUN go get github.com/nsf/gocode
RUN go get github.com/bradfitz/goimports

ADD . /go/src/github.com/b3log/wide
WORKDIR /go/src/github.com/b3log/wide
RUN go get
RUN go build

RUN cp -r . /root/wide
WORKDIR /root/wide
RUN rm -rf /go/bin /go/pkg /go/src
RUN mv ./hello /go/src/

ENV GOROOT /usr/src/go

EXPOSE 7070