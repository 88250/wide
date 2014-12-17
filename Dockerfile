FROM golang:latest
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /wide/gogogo/src/github.com/b3log/wide

RUN useradd wide && chown -R wide:wide /wide && wide_runner

USER wide

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN go get -v github.com/88250/ide_stub github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go get -v && go build -v

EXPOSE 7070