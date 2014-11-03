#!/usr/bin/env bash

# Wide package tool.
# 
# Command: 
#  ./pkg.bash ${version} ${target}
# Example:
#  ./pkg.bash 1.0.1 /home/daniel/1.0.1/

ver=$1
target=$2
list="conf data doc i18n static views README.md LICENSE"

mkdir ${target}

echo version=${ver}
echo target=${target}
echo 

## darwin
os=darwin

export GOOS=${os}
export GOARCH=386
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs
rm -f wide

export GOOS=${os}
export GOARCH=amd64
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs
rm -f wide

## linux
os=linux

export GOOS=${os}
export GOARCH=386
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs
rm -f wide

export GOOS=${os}
export GOARCH=amd64
go build
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} wide --exclude-vcs
rm -f wide

## windows
os=windows

export GOOS=${os}
export GOARCH=386
go build
zip -r ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} wide.exe
rm -f wide.exe

export GOOS=${os}
export GOARCH=amd64
go build
zip -r ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} wide.exe
rm -f wide.exe
