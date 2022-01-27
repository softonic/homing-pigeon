FROM golang:1.17-buster AS build

RUN apt-get update \
    && apt-get install -y protobuf-compiler

WORKDIR /go/src/github.com/softonic/homing-pigeon

COPY . .

RUN make build &&\
    ln /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

FROM scratch

COPY --from=build /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

ENTRYPOINT ["/homing-pigeon", "-logtostderr"]
