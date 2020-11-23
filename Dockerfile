FROM golang:1.13-buster AS build

RUN apt-get update \
    && apt-get install -y protobuf-compiler

WORKDIR /go/src/github.com/softonic/homing-pigeon

COPY . .

RUN make build &&\
    cp /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

FROM scratch

COPY --from=build /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

ENTRYPOINT ["/homing-pigeon", "-logtostderr"]
