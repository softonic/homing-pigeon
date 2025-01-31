FROM golang:1.23-bookworm AS build

RUN apt-get update \
    && apt-get install -y protobuf-compiler

RUN groupadd --gid 1000 app && useradd -u 1000 -g app app

WORKDIR /go/src/github.com/softonic/homing-pigeon

COPY . .

RUN make build &&\
    ln /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

FROM scratch

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --chown=app:app --from=build /go/src/github.com/softonic/homing-pigeon/bin/homing-pigeon /

USER app

ENTRYPOINT ["/homing-pigeon", "-logtostderr"]
