package main

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/middleware"
	"github.com/softonic/homing-pigeon/pkg/readers"
	"github.com/softonic/homing-pigeon/pkg/writers"
	"k8s.io/klog"
	"os"
	"strconv"
)

func main() {
	klog.InitFlags(nil)

	bufLen, err := strconv.Atoi(os.Getenv("MESSAGE_BUFFER_LENGTH"))
	if err != nil {
		bufLen = 0
	}
	inputChannel := make(chan messages.Message, bufLen)
	ackChannel := make(chan messages.Ack, bufLen)
	outputChannel := make(chan messages.Message, bufLen)

	reader, err := readers.NewAMQPReader(inputChannel, ackChannel)
	if err != nil {
		panic(err)
	}

	middleware := middleware.NewMiddlewareManager(inputChannel, outputChannel)
	writer, err := writers.NewElasticsearchWriter(outputChannel, ackChannel)
	if err != nil {
		panic(err)
	}
	go reader.Start()
	go middleware.Start()
	writer.Start()
}
