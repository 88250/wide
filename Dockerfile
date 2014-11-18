FROM golang:latest

RUN go get github.com/88250/ide_stub
RUN go get github.com/nsf/gocode
RUN go get github.com/bradfitz/goimports

ADD . /go/src/github.com/b3log/wide

WORKDIR /go/src/github.com/b3log/wide

RUN go get
RUN go build

ENV GOROOT /usr/src/go

EXPOSE 7070