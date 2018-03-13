FROM golang:latest
MAINTAINER Liang Ding <d@b3log.org>

ENV GOROOT /usr/local/go

RUN apt-get update && apt-get install bzip2 zip unzip && cp -r /usr/local/go /usr/local/gobt
ENV GOROOT_BOOTSTRAP=/usr/local/gobt

ADD . /wide/gogogo/src/github.com/b3log/wide
ADD vendor/* /go/src/
RUN go install github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports

RUN useradd wide && useradd runner

ENV GOPATH /wide/gogogo

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go build -v

EXPOSE 7070
