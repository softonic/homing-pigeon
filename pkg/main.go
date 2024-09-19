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
	msgChBeforeMiddleware := make(chan messages.Message, bufLen)
	msgChAfterMiddleware := make(chan messages.Message, bufLen)

	bufLen = GetBufferLength("ACK_BUFFER_LENGTH")
	ackChBeforeMiddleware := make(chan messages.Message, bufLen)
	ackChAfterMiddleware := make(chan messages.Message, bufLen)

	reader, err := readers.NewReader(msgChBeforeMiddleware, ackChAfterMiddleware)
	if err != nil {
		panic(err)
	}

	writer, err := writers.NewWriter(msgChAfterMiddleware, ackChBeforeMiddleware)
	if err != nil {
		panic(err)
	}

	requestMiddleware := middleware.NewMiddlewareManager(
		msgChBeforeMiddleware,
		msgChAfterMiddleware,
		helpers.GetEnv("REQUEST_MIDDLEWARES_SOCKET", ""),
	)

	responseMiddleware := middleware.NewMiddlewareManager(
		ackChBeforeMiddleware,
		ackChAfterMiddleware,
		helpers.GetEnv("RESPONSE_MIDDLEWARES_SOCKET", ""),
	)

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
