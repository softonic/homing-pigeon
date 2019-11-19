build:
	dingo -src="./pkg/services" -dest="./pkg/generatedServices"
	go build -o bin/homing-pigeon pkg/main.go
