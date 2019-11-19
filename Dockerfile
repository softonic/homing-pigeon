FROM golang:1.13-buster AS build

COPY ./ /go/src/github.com/softonic/homing-pigeon

RUN cd /go/src/github.com/softonic/homing-pigeon && make build

FROM scratch

COPY --from=build /go/src/github.com/softonic/homing-pigeon/bin /

ENTRYPOINT ["/homing-pigeon"]