package main

import (
	"context"
	pb "github.com/softonic/homing-pigeon/middleware"

	"google.golang.org/grpc"
	"log"
	"net"
)

type MiddlewareServer struct {
	pb.UnimplementedMiddlewareServer
}

func (*MiddlewareServer) Transform(ctx context.Context, req *pb.Data) (*pb.Data, error) {
	log.Printf("Processing %v", *req)
	return req, nil
}

func main() {
	lis, err := net.Listen("unix", "/tmp/hp")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	pb.RegisterMiddlewareServer(grpcServer, &MiddlewareServer{})

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
