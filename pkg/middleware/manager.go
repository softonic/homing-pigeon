package middleware

import (
	"context"
	"time"

	"github.com/softonic/homing-pigeon/pkg/helpers"
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

func (m *MiddlwareManager) Start() {
	if m.isMiddlewareNotAvailable() {
		klog.V(1).Infof("Middlewares not available")
		for message := range m.InputChannel {
			m.OutputChannel <- message
		}
	}

	klog.V(1).Infof("Middlewares available")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

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

		data, err := client.Handle(context.Background(), &proto.Data{
			Body: message.Body,
		})
		if err != nil {
			klog.Errorf("What happens!? %v", err)
		}

		elapsed := time.Since(start)
		klog.V(5).Infof("Middlewares took %s", elapsed)

		message.Body = data.GetBody()

		m.OutputChannel <- message
	}
}

func (m *MiddlwareManager) isMiddlewareNotAvailable() bool {
	return m.MiddlewareAddress == ""
}

func NewMiddlewareManager(msgCh1 chan messages.Message, msgCh2 chan messages.Message) *MiddlwareManager {
	return &MiddlwareManager{
		InputChannel:      msgCh1,
		OutputChannel:     msgCh2,
		MiddlewareAddress: helpers.GetEnv("MIDDLEWARES_SOCKET", ""),
	}
}
