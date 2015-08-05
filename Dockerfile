FROM golang:cross
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /wide/gogogo/src/github.com/b3log/wide

RUN tar zxf /wide/gogogo/src/github.com/b3log/wide/docker-deps/golang.org.tar.gz -C /wide/gogogo/src/
RUN tar zxf /wide/gogogo/src/github.com/b3log/wide/docker-deps/ide_stud.tar.gz -C /wide/gogogo/src/github.com/88250/

RUN useradd wide && useradd runner

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN go get -v github.com/88250/ide_stub github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go get -v && go build -v

EXPOSE 7070