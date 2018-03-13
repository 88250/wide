FROM golang:latest
MAINTAINER Liang Ding <d@b3log.org>

ADD . /go/src/github.com/b3log/wide
ADD vendor/ /go/src/
RUN go install github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports

RUN useradd wide && useradd runner

WORKDIR /go/src/github.com/b3log/wide
RUN go build -v

EXPOSE 7070
