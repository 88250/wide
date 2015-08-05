FROM golang:cross
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /wide/gogogo/src/github.com/b3log/wide
RUN tar zxf /wide/gogogo/src/github.com/b3log/wide/golang.org.tar.gz /wide/gogogo/src/

RUN useradd wide && useradd runner

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN github.com/88250/ide_stub github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go get -v && go build -v

EXPOSE 7070