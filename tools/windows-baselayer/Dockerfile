FROM golang:1.15-buster

RUN apt update \
    && apt install -y libhivex-dev libhivex-bin libwin-hivex-perl

COPY . /src

WORKDIR /src

VOLUME /go/pkg/mod
