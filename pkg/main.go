package main

import (
	"github.com/softonic/homing-pigeon/pkg/generatedServices/dic"
	"github.com/sarulabs/dingo"
)

func main() {
	container, err := dic.NewContainer(dingo.App)
	if err != nil {
		panic(err)
	}
	amqpReader := container.GetAmqpReader()
	elasticsearchWriter := container.GetElasticsearchWriter()

	go amqpReader.Start()
	elasticsearchWriter.Start()
}
