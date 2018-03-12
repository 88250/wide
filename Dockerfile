FROM golang:1.8.3
MAINTAINER Liang Ding <d@b3log.org>

ENV GOROOT /usr/local/go

RUN apt-get update && apt-get install bzip2 zip unzip && cp -r /usr/local/go /usr/local/gobt
ENV GOROOT_BOOTSTRAP=/usr/local/gobt

ADD . /wide/gogogo/src/github.com/b3log/wide

RUN unzip /wide/gogogo/src/github.com/b3log/wide/deps/golang.org.zip -d /wide/gogogo/src/\
 && unzip /wide/gogogo/src/github.com/b3log/wide/deps/github.com.zip -d /wide/gogogo/src/\
 && useradd wide && useradd runner

ENV GOPATH /wide/gogogo

RUN go build github.com/go-fsnotify/fsnotify\
 && go build github.com/gorilla/sessions\
 && go build github.com/gorilla/websocket\
 && go install github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go build -v

EXPOSE 7070
