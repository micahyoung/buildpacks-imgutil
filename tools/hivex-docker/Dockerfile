FROM golang:1.15-buster

RUN apt update \
    && apt install -y libhivex-dev libhivex-bin libwin-hivex-perl \
    && go get -u github.com/kyoh86/richgo

ENV GOCMD=richgo

COPY . /src

WORKDIR /src

VOLUME /go/pkg/mod
