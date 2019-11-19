package main

import (
	"github.com/sarulabs/dingo"
	"github.com/softonic/homing-pigeon/pkg/generatedServices/dic"
)

func main() {
	container, err := dic.NewContainer(dingo.App)
	if err != nil {
		panic(err)
	}
	amqpReader := container.GetAmqpReader()
	writer := container.GetWriter()

	go amqpReader.Start()
	writer.Start()
}
