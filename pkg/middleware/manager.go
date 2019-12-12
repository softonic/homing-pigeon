package middleware

import (
	"context"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"time"
)

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
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(m.MiddlewareAddress, opts...)
	if err != nil {
		klog.Errorf("fail to dial: %v", err)
	}

	klog.V(1).Infof("Middlewares connected")

	defer conn.Close()
	client := proto.NewMiddlewareClient(conn)

	for message := range m.InputChannel {
		klog.V(5).Infof("Sending message to proto", )
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
