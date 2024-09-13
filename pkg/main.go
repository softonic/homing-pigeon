package main

import (
	"os"
	"strconv"

	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/middleware"
	"github.com/softonic/homing-pigeon/pkg/readers"
	"github.com/softonic/homing-pigeon/pkg/writers"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)

	bufLen := GetBufferLength("MESSAGE_BUFFER_LENGTH")
	msgCh1 := make(chan messages.Message, bufLen)
	msgCh2 := make(chan messages.Message, bufLen)

	bufLen = GetBufferLength("ACK_BUFFER_LENGTH")
	ackCh1 := make(chan messages.Message, bufLen)
	ackCh2 := make(chan messages.Message, bufLen)

	reader, err := readers.NewReader(msgCh1, ackCh1)
	if err != nil {
		panic(err)
	}

	writer, err := writers.NewWriter(msgCh2, ackCh2)
	if err != nil {
		panic(err)
	}

	requestMiddleWareAddress := helpers.GetEnv("REQUEST_MIDDLEWARES_SOCKET", "")
	requestMiddleware := middleware.NewMiddlewareManager(msgCh1, msgCh2, requestMiddleWareAddress)

	responseMiddleWareAddress := helpers.GetEnv("RESPONSE_MIDDLEWARES_SOCKET", "")
	responseMiddleware := middleware.NewMiddlewareManager(ackCh2, ackCh1, responseMiddleWareAddress)

	go reader.Start()
	go requestMiddleware.Start()
	go responseMiddleware.Start()
	writer.Start()
}

// GetBufferLength returns the buffer length for bufferKey env
func GetBufferLength(bufferKey string) int {
	bufLen, err := strconv.Atoi(os.Getenv(bufferKey))
	if err != nil {
		bufLen = 0
	}

	return bufLen
}
