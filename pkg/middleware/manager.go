package middleware

import (
	"context"
	"github.com/softonic/homing-pigeon/middleware"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"google.golang.org/grpc"
	"log"
	"time"
)

type MiddlwareManager struct {
	InputChannel      <-chan messages.Message
	OutputChannel     chan<- messages.Message
	MiddlewareAddress string
}

func (m *MiddlwareManager) Start() {
	if m.isMiddlewareNotAvailable() {
		log.Printf("Middlewares not available")
		for message := range m.InputChannel {
			m.OutputChannel <- message
		}
	}

	log.Printf("Middlewares available")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(m.MiddlewareAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	log.Print("Middlewares connected")

	defer conn.Close()
	client := middleware.NewMiddlewareClient(conn)

	for message := range m.InputChannel {
		log.Printf("Sending message to middleware",)
		start := time.Now()

		data, err := client.Handle(context.Background(), &middleware.Data{
			Body: message.Body,
		})
		if err != nil {
			log.Fatalf("What happens!? %v", err)
		}

		elapsed := time.Since(start)
		log.Printf("Middlewares took %s", elapsed)

		message.Body = data.GetBody()

		m.OutputChannel <- message
	}
}

func (m *MiddlwareManager) isMiddlewareNotAvailable() bool {
	return m.MiddlewareAddress == ""
}
