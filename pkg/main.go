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

	bufLen := GetBufferLength("MESSAGE_BUFFER_LENGTH")
	msgCh1 := make(chan messages.Message, bufLen)
	msgCh2 := make(chan messages.Message, bufLen)

	bufLen = GetBufferLength("ACK_BUFFER_LENGTH")
	ackCh1 := make(chan messages.Ack, bufLen)
	ackCh2 := make(chan messages.Ack, bufLen)

	reader, err := readers.NewReader(msgCh1, ackCh2)
	if err != nil {
		panic(err)
	}

	writer, err := writers.NewWriter(msgCh2, ackCh1)
	if err != nil {
		panic(err)
	}

	incomingMiddleware := middleware.NewMiddlewareManager(msgCh1, msgCh2, ackCh1, ackCh2)

	go reader.Start()
	go incomingMiddleware.HandleInput()
	go incomingMiddleware.HandleOutput()
	writer.Start()
}

func GetBufferLength(bufferKey string) int {
	bufLen, err := strconv.Atoi(os.Getenv(bufferKey))
	if err != nil {
		bufLen = 0
	}

	return bufLen
}
