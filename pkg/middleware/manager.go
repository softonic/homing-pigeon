package middleware

import (
	"context"
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog"
)

// @TODO Tests missing
type MiddlwareManager struct {
	InputChannel      <-chan messages.Message
	OutputChannel     chan<- messages.Message
	MiddlewareAddress string
}

// Start starts the middleware manager.
func (m *MiddlwareManager) Start() {
	if m.isMiddlewareNotAvailable() {
		klog.V(1).Infof("Middlewares not available")
		for message := range m.InputChannel {
			m.OutputChannel <- message
		}
	}

	klog.V(1).Infof("Middlewares available")

	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(defaultRetryPolicy))

	conn, err := grpc.NewClient(m.MiddlewareAddress, opts...)
	if err != nil {
		klog.Errorf("fail to dial: %v", err)
	}

	klog.V(1).Infof("Middlewares connected")

	defer conn.Close()
	client := proto.NewMiddlewareClient(conn)

	for message := range m.InputChannel {
		klog.V(5).Infof("Sending message to proto")
		start := time.Now()

		// wait for ready up to 12 seconds (including retries)
		// to handle service discontinuity (external middlewares) or startup order
		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 31*time.Second)
		handleData, err := client.Handle(ctxTimeout, &proto.Data{
			Body: message.Body,
		}, grpc.WaitForReady(true))
		cancelTimeout()
		if err != nil {
			klog.Errorf("Error calling middleware %v", err)
		} else {
			message.Body = handleData.GetBody()
		}

		elapsed := time.Since(start)
		klog.V(5).Infof("Middlewares took %s", elapsed)

		m.OutputChannel <- message
	}
}

func (m *MiddlwareManager) isMiddlewareNotAvailable() bool {
	return m.MiddlewareAddress == ""
}

// NewMiddlewareManager creates a new instance of MiddlwareManager.
func NewMiddlewareManager(inputChannel chan messages.Message, outputChannel chan messages.Message, address string) *MiddlwareManager {
	return &MiddlwareManager{
		InputChannel:      inputChannel,
		OutputChannel:     outputChannel,
		MiddlewareAddress: address,
	}
}
