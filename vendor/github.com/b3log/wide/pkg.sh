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

export GOOS=${os}
export GOARCH=amd64
echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zcf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

export GOOS=${os}
export GOARCH=386
echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zcf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

## linux
os=linux

export GOOS=${os}
export GOARCH=amd64
echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zcf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

export GOOS=${os}
export GOARCH=386
echo wide-${ver}-${GOOS}-${GOARCH}.tar.gz
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
tar zcf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} gotools gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide gotools gocode

## windows
os=windows

export GOOS=${os}
export GOARCH=amd64
echo wide-${ver}-${GOOS}-${GOARCH}.zip
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
zip -r -q ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} gotools.exe gocode.exe wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe gotools.exe gocode.exe

export GOOS=${os}
export GOARCH=386
echo wide-${ver}-${GOOS}-${GOARCH}.zip
go build
go build github.com/visualfc/gotools
go build github.com/nsf/gocode
zip -r -q ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} gotools.exe gocode.exe wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe gotools.exe gocode.exe
