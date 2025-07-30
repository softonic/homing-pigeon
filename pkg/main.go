package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	// Third-party packages, with a blank line separator
	_ "go.uber.org/automaxprocs"
	"k8s.io/klog"

	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/middleware"
	"github.com/softonic/homing-pigeon/pkg/readers"
	"github.com/softonic/homing-pigeon/pkg/writers"
)

func main() {
	// Setup graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals (SIGTERM from Kubernetes, SIGINT from Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c      // Wait for shutdown signal
		cancel() // Trigger graceful shutdown
	}()

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

	batchSize := helpers.GetIntEnv("MIDDLEWARE_BATCH_SIZE", 50)
	batchTimeout := time.Duration(helpers.GetIntEnv("MIDDLEWARE_BATCH_TIMEOUT_MS", 100)) * time.Millisecond

	requestMiddleware := middleware.NewMiddlewareManager(
		msgChBeforeMiddleware,
		msgChAfterMiddleware,
		helpers.GetEnv("REQUEST_MIDDLEWARES_SOCKET", ""),
		batchSize,
		batchTimeout,
	)

	responseMiddleware := middleware.NewMiddlewareManager(
		ackChBeforeMiddleware,
		ackChAfterMiddleware,
		helpers.GetEnv("RESPONSE_MIDDLEWARES_SOCKET", ""),
		batchSize,
		batchTimeout,
	)

	go reader.Start(ctx)
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
