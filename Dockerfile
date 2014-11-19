FROM golang:latest
MAINTAINER Liang Ding <dl88250@gmail.com>

RUN ./docker.sh

ENV GOROOT /usr/src/go

EXPOSE 7070