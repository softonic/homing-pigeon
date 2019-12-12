package middleware

import (
	"context"
	pb "github.com/softonic/homing-pigeon/middleware"
	. "github.com/softonic/homing-pigeon/pkg/helpers"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"path/filepath"
)

type Base struct {
	client *pb.MiddlewareClient
	pb.UnimplementedMiddlewareServer
}

func (b *Base) Next(req *pb.Data) (*pb.Data, error) {
	resp := req
	if b.client != nil {
		var err error
		resp, err = (*b.client).Handle(context.Background(), req)
		return nil, err
	}

	return resp, nil
}

func (b *Base) Listen(middleware pb.MiddlewareServer) {
	lis, err := net.Listen("unix", b.getInputSocket())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	conn, client := b.getOutputGrpc()
	defer conn.Close()

	b.client = client

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	pb.RegisterMiddlewareServer(grpcServer, middleware)

	log.Print("Start listening")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Base) getInputSocket() string {
	socket := GetEnv("IN_SOCKET", "")

	err := os.RemoveAll(socket)
	if err != nil {
		log.Fatalf("Failed to remove socket: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(socket), 0775)
	if err != nil {
		log.Fatalf("Error creating socket directory: %v", err)
	}
	return socket
}

func (b *Base) getOutputGrpc() (*grpc.ClientConn, *pb.MiddlewareClient) {
	nextSocketAddr := GetEnv("OUT_SOCKET", "")
	if nextSocketAddr != "" {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithBlock())

		conn, err := grpc.Dial(nextSocketAddr, opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}

		log.Print("Connected to the next middleware")

		client := pb.NewMiddlewareClient(conn)
		return conn, &client
	}

	return nil, nil
}
