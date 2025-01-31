TAG ?= dev

generate-proto:
	protoc -I proto/ proto/middleware.proto --go_out=plugins=grpc:proto
dep:
	go get -u github.com/rakyll/gotest
	go get -u github.com/vektra/mockery/.../
	go get -u github.com/golang/protobuf/proto
	go install github.com/golang/protobuf/protoc-gen-go@v1.3.2
	go get -u google.golang.org/grpc
	go mod download
build: dep generate-proto
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/homing-pigeon pkg/main.go
stress-build: dep generate-proto
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/stress-pigeon pkg/stress/main.go
docker-build:
	docker buildx create --name hp-image-builder --driver docker-container --bootstrap 2> /dev/null || true
	docker buildx use hp-image-builder
	docker buildx build \
		--pull \
		--push \
		-f ./Dockerfile \
		. \
		--platform linux/arm64,linux/amd64 \
		-t softonic/homing-pigeon:${TAG}
mock:
	mockery --name=WriteAdapter -r
	mockery --name=Channel -r
	mockery --name=Connection -r
test:
	gotest -race -count=1 ./... -v
