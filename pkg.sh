#!/bin/bash

# Wide package tool.
# 
# Command: 
#  ./pkg.sh ${version} ${target}
# Example:
#  ./pkg.sh 1.0.1 /home/daniel/1.0.1/

ver=$1
target=$2
list="conf doc i18n static views README.md LICENSE"

mkdir -p ${target}

echo version=${ver}
echo target=${target}

## darwin
os=darwin

export GOOS=${os}
export GOARCH=386
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs --exclude conf/*.go --exclude i18n/*.go
rm -f wide

export GOOS=${os}
export GOARCH=amd64
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs --exclude conf/*.go --exclude i18n/*.go
rm -f wide

## linux
os=linux

export GOOS=${os}
export GOARCH=386
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs --exclude conf/*.go --exclude i18n/*.go
rm -f wide

export GOOS=${os}
export GOARCH=amd64
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs --exclude conf/*.go --exclude i18n/*.go
rm -f wide

## windows
os=windows

export GOOS=${os}
export GOARCH=386
go build
zip -r ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe

export GOOS=${os}
export GOARCH=amd64
go build
zip -r ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe
