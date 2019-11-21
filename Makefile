dep: go.mod go.sum
	go get -u github.com/sarulabs/dingo/dingo
	go get -u github.com/vektra/mockery/.../
	go mod download
build: dep
	dingo -src="./pkg/services" -dest="./pkg/generatedServices"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/homing-pigeon pkg/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/stress-pigeon pkg/stress.go
docker-build:
	docker build -t softonic/homing-pigeon:dev .
mock:
	mockery -name=BulkClient -recursive
	mockery -name=WriteAdapter -recursive
test:
	go test ./... -v