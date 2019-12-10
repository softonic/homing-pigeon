package middleware

import (
	"context"
	transformer "github.com/softonic/homing-pigeon/middleware"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"google.golang.org/grpc"
	"log"
	"time"
)

type Middlware struct {
	InputChannel     <-chan messages.Message
	OutputChannel    chan<- messages.Message
	MiddlewareSocket string
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

	conn, err := grpc.Dial(m.MiddlewareSocket, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()
	client := transformer.NewMiddlewareClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for message := range m.InputChannel {
		data, err := client.Transform(ctx, &transformer.Data{
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
	return m.MiddlewareSocket == ""
}
