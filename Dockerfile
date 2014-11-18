FROM golang:latest

ADD . /go/src/github.com/b3log/wide

WORKDIR /go/src/github.com/b3log/wide

RUN go get
RUN go build

ENV GOROOT /usr/src/go

EXPOSE 7070