#!/usr/bin/env bash

# Wide package tool.
# 
# See https://github.com/gobuild/gobuild3/tree/master/packer

ver=$1

echo $ver

./packer --os darwin --arch 386 -o wide-$ver-darwin-386.tar.gz
./packer --os darwin --arch amd64 -o wide-$ver-darwin-amd64.tar.gz

./packer --os linux --arch 386 -o wide-$ver-linux-386.tar.gz
./packer --os linux --arch amd64 -o wide-$ver-linux-amd64.tar.gz

./packer --os windows --arch 386 -o wide-$ver-windows-386.zip
./packer --os windows --arch amd64 -o wide-$ver-windows-amd64.zip
