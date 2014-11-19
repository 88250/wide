#!/bin/bash

# Make docker image.

go get github.com/88250/ide_stub
go get github.com/nsf/gocode
go get github.com/bradfitz/goimports

cp -r . /go/src/github.com/b3log/wide
cd /go/src/github.com/b3log/wide
go get
go build
cp -r . /root/wide

cd /root/wide
rm -rf /go/pkg /go/src
mv ./hello /go/src/hello
