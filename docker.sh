#!/bin/bash

# Make docker image.

go get
go build
cp -r . /root/wide

go get github.com/88250/ide_stub
go get github.com/nsf/gocode
go get github.com/bradfitz/goimports

cd /root/wide
rm -rf /go/pkg /go/src
mv ./hello /go/src/
