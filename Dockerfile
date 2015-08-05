FROM golang:cross
MAINTAINER Liang Ding <dl88250@gmail.com>

ADD . /wide/gogogo/src/github.com/b3log/wide

RUN tar zxf /wide/gogogo/src/github.com/b3log/wide/deps/golang.org.tar.gz -C /wide/gogogo/src/
RUN tar zxf /wide/gogogo/src/github.com/b3log/wide/deps/github.com.tar.gz -C /wide/gogogo/src/

RUN useradd wide && useradd runner

ENV GOROOT /usr/src/go
ENV GOPATH /wide/gogogo

RUN go build github.com/go-fsnotify/fsnotify 
RUN go build github.com/gorilla/sessions 
RUN go build github.com/gorilla/websocket

RUN go install github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports

WORKDIR /wide/gogogo/src/github.com/b3log/wide
RUN go build -v

EXPOSE 7070