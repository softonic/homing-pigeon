package main

import (
	"bitbucket.org/softonic-development/homming-pidgeon/pkg/generatedServices/dic"
	"github.com/sarulabs/dingo"
)

func main() {
	container, err := dic.NewContainer(dingo.App)
	if err != nil {
		panic(err)
	}
	inChannel = make(chan string)
	outChannel = make(chan string)
	amqpReader := container.GetAmqpReader(inChannel, outChannel)
	elasticsearchWriter := container.GetElasticsearchWriter(inChannel, outChannel)


}
