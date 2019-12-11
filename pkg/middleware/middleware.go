package middleware

import (
	"context"
	transformer "github.com/softonic/homing-pigeon/middleware"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"google.golang.org/grpc"
	"log"
)

type Middlware struct {
	InputChannel      <-chan messages.Message
	OutputChannel     chan<- messages.Message
	MiddlewareAddress string
}

func (m *Middlware) Start() {
	if m.isMiddlewareNotAvailable() {
		for message := range m.InputChannel {
			m.OutputChannel <- message
		}
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(m.MiddlewareAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()
	client := transformer.NewMiddlewareClient(conn)

	for message := range m.InputChannel {
		data, err := client.Transform(context.Background(), &transformer.Data{
			Body: message.Body,
		})

		if err != nil {
			log.Fatalf("What happens!? %v", err)
		}
		message.Body = data.GetBody()

		m.OutputChannel <- message
	}
}

func (m *Middlware) isMiddlewareNotAvailable() bool {
	return m.MiddlewareAddress == ""
}
