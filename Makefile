dep: go.mod go.sum
	go get -u github.com/sarulabs/dingo/dingo
	go mod download
build: dep
	dingo -src="./pkg/services" -dest="./pkg/generatedServices"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o bin/homing-pigeon pkg/main.go
docker-build:
	docker build -t softonic/homing-pigeon:dev .