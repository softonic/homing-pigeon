FROM golang:1.13-buster AS build

RUN apt-get update \
    && apt-get install -y protobuf-compiler

COPY ./ /go/src/github.com/softonic/homing-pigeon

RUN cd /go/src/github.com/softonic/homing-pigeon && make build

FROM scratch

COPY --from=build /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

ENTRYPOINT ["/homing-pigeon", "-logtostderr"]