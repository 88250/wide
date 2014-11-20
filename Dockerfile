FROM golang:latest
MAINTAINER Liang Ding <dl88250@gmail.com>

RUN useradd wide && mkdir -p /wide/gogogo/ && chown -R wide:wide /wide
USER wide

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN go get github.com/88250/ide_stub github.com/nsf/gocode github.com/bradfitz/goimports

ADD . /wide/gogogo/src/github.com/b3log/wide
WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go get && go build

EXPOSE 7070