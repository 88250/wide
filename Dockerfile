FROM golang:latest
MAINTAINER Liang Ding <d@b3log.org>

ENV GOROOT /usr/local/go

RUN apt-get update && apt-get install bzip2 zip unzip && cp -r /usr/local/go /usr/local/gobt
ENV GOROOT_BOOTSTRAP=/usr/local/gobt

ADD . /wide/gogogo/src/github.com/b3log/wide
ADD vendor/* /wide/gogogo/src/

RUN useradd wide && useradd runner

ENV GOPATH /wide/gogogo

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go build -v

EXPOSE 7070
