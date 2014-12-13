FROM golang:latest
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /wide/gogogo/src/github.com/b3log/wide

RUN useradd wide && chown -R wide:wide /wide
USER wide

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN go get -v github.com/88250/ide_stub github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go get -v && go build -v

RUN ln -sf /dev/stdout /var/log/wide/out.log
RUN ln -sf /dev/stderr /var/log/wide/err.log

EXPOSE 7070