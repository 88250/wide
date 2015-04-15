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
export GOARCH=386
go build
go build github.com/88250/ide_stub
go build github.com/nsf/gocode
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/ide_stub .
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/gocode .
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} ide_stub gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide ide_stub gocode

export GOOS=${os}
export GOARCH=amd64
go build
go build github.com/88250/ide_stub
go build github.com/nsf/gocode
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/ide_stub .
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/gocode .
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} ide_stub gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide ide_stub gocode

## linux
os=linux

export GOOS=${os}
export GOARCH=386
go build
go build github.com/88250/ide_stub
go build github.com/nsf/gocode
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/ide_stub .
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/gocode .
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} ide_stub gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide ide_stub gocode

export GOOS=${os}
export GOARCH=amd64
go build
go build github.com/88250/ide_stub
go build github.com/nsf/gocode
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/ide_stub .
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/gocode .
tar zcvf ${target}/wide-${ver}-${GOOS}-${GOARCH}.tar.gz ${list} ide_stub gocode wide --exclude-vcs --exclude='conf/*.go' --exclude='i18n/*.go'
rm -f wide ide_stub gocode

## windows
os=windows

export GOOS=${os}
export GOARCH=386
go build
go build github.com/88250/ide_stub
go build github.com/nsf/gocode
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/ide_stub.exe .
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/gocode.exe .
zip -r ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} ide_stub.exe gocode.exe wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe ide_stub.exe gocode.exe

export GOOS=${os}
export GOARCH=amd64
go build
go build github.com/88250/ide_stub
go build github.com/nsf/gocode
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/ide_stub.exe .
cp ${GOPATH}/bin/${GOOS}_${GOARCH}/gocode.exe .
zip -r ${target}/wide-${ver}-${GOOS}-${GOARCH}.zip ${list} ide_stub.exe gocode.exe wide.exe --exclude=conf/*.go --exclude=i18n/*.go
rm -f wide.exe ide_stub.exe gocode.exe

