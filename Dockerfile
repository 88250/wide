FROM golang:cross
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /wide/gogogo/src/github.com/b3log/wide

RUN useradd wide && useradd runner

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN go get -v golang.org/x/tools/cmd/vet github.com/88250/ide_stub github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go get -v && go build -v

EXPOSE 7070