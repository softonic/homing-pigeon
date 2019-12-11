FROM golang:1.13-buster AS build

COPY ./ /go/src/github.com/softonic/homing-pigeon

RUN cd /go/src/github.com/softonic/homing-pigeon && make build

FROM debian:buster

COPY --from=build /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

ENTRYPOINT ["/homing-pigeon"]