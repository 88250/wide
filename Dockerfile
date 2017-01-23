FROM golang:1.7.4
MAINTAINER Liang Ding <dl88250@gmail.com>

ENV GOLANG_CROSSPLATFORMS \
	darwin/386 darwin/amd64 \
	dragonfly/386 dragonfly/amd64 \
	freebsd/386 freebsd/amd64 freebsd/arm \
	linux/386 linux/amd64 linux/arm \
	nacl/386 nacl/amd64p32 nacl/arm \
	netbsd/386 netbsd/amd64 netbsd/arm \
	openbsd/386 openbsd/amd64 \
	plan9/386 plan9/amd64 \
	solaris/amd64 \
	windows/386 windows/amd64
ENV GOARM 5

RUN cd /usr/local/go/src \
	&& set -ex \
	&& for platform in $GOLANG_CROSSPLATFORMS; do \
		GOOS=${platform%/*} \
		GOARCH=${platform##*/} \
		./make.bash --no-clean 2>&1; \
	done

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