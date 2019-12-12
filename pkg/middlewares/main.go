package main

import (
	"context"
	pb "github.com/softonic/homing-pigeon/middleware"
	"github.com/softonic/homing-pigeon/pkg/middleware"
	"log"
)

type PassthroughMiddleware struct {
	middleware.Base
}

func (m *PassthroughMiddleware) Handle(ctx context.Context, req *pb.Data) (*pb.Data, error) {

	// Do things with the INPUT data
	log.Printf("Pre-Processing %v", *req)

	// Send data to the next middleware and got the response
	resp, err := m.Next(req)

	// Do things with the OUTPUT data
	log.Printf("Post-Processing %v", *resp)

	return resp, err
}

func main() {
	middleware := &PassthroughMiddleware{}
	middleware.Listen(middleware)
}
