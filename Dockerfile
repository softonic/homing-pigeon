FROM golang:1.23-bookworm AS build

RUN apt-get update \
    && apt-get install -y protobuf-compiler

WORKDIR /go/src/github.com/softonic/homing-pigeon

COPY . .

RUN make build &&\
    ln /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

FROM scratch

COPY --from=build /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

ENTRYPOINT ["/homing-pigeon", "-logtostderr"]
