TAG ?= dev
PROTOC_VERSION = 3.14.0

install-protoc:
	curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	exit
	sudo unzip -o protoc-$(PROTOC_VERSION)-linux-x86_64.zip -d /usr/local bin/protoc
	sudo unzip -o protoc-$(PROTOC_VERSION)-linux-x86_64.zip -d /usr/local 'include/*'
	sudo chmod +x /usr/local/bin/protoc
	rm -f protoc-$(PROTOC_VERSION)-linux-x86_64.zip
generate-proto: install-protoc
	protoc -I proto/ proto/middleware.proto --go_out=plugins=grpc:proto
dep:
	go get -u github.com/rakyll/gotest
	go get -u github.com/sarulabs/dingo/dingo
	go get -u github.com/vektra/mockery/.../
	go get -u github.com/golang/protobuf/proto
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u google.golang.org/grpc
	go mod download
build: dep generate-proto
	dingo -src="./pkg/services" -dest="./pkg/generatedServices"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/homing-pigeon pkg/main.go
stress-build: dep generate-proto
	dingo -src="./pkg/services" -dest="./pkg/generatedServices"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/stress-pigeon pkg/stress/main.go
docker-build:
	docker build -t softonic/homing-pigeon:${TAG} .
mock:
	mockery -name=WriteAdapter -recursive
	mockery -name=Channel -recursive
	mockery -name=Connection -recursive
test:
	gotest -race -count=1 ./... -v
