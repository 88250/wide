#!/bin/bash

# Wide package tool.
#
# Command: 
#  ./pkg.sh ${version} ${target}
# Example:
#  ./pkg.sh 1.0.0 /home/daniel/1.0.0/

ver=$1
target=$2
list="conf doc i18n static views README.md TERMS.md LICENSE"

mkdir -p ${target}

echo version=${ver}
echo target=${target}

## darwin
os=darwin

echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
export GOOS=${os}
export GOARCH=386
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
export GOOS=${os}
export GOARCH=amd64
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

## linux
os=linux

echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
export GOOS=${os}
export GOARCH=386
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
export GOOS=${os}
export GOARCH=amd64
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

## windows
os=windows

echo wide-${ver}-${GOOS}-${GOARCH}.zip
export GOOS=${os}
export GOARCH=386
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
zip -rq ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} gotools.exe gocode.exe wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe gotools.exe gocode.exe

echo wide-${ver}-${GOOS}-${GOARCH}.zip
export GOOS=${os}
export GOARCH=amd64
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
zip -rq ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} gotools.exe gocode.exe wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe gotools.exe gocode.exe

