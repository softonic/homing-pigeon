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

	log.Printf("Pre-Processing %v", *req)
	// Do things
	//---------------------
	resp, err := m.Next(req)

	log.Printf("Post-Processing %v", *resp)
	// Do things

	return resp, err
}

func main() {
	middleware := PassthroughMiddleware{}
	middleware.Listen()
}
