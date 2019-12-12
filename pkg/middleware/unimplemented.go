package middleware

import (
	"context"
	. "github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"path/filepath"
)

type UnimplementedMiddleware struct {
	client *proto.MiddlewareClient
	proto.UnimplementedMiddlewareServer
}

func (b *UnimplementedMiddleware) Next(req *proto.Data) (*proto.Data, error) {
	resp := req
	if b.client != nil {
		var err error
		resp, err = (*b.client).Handle(context.Background(), req)
		return nil, err
	}

	return resp, nil
}

func (b *UnimplementedMiddleware) Listen(middleware proto.MiddlewareServer) {
	lis, err := net.Listen("unix", b.getInputSocket())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	conn, client := b.getOutputGrpc()
	defer conn.Close()

	b.client = client

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	proto.RegisterMiddlewareServer(grpcServer, middleware)

	log.Print("Start listening")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *UnimplementedMiddleware) getInputSocket() string {
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

func (b *UnimplementedMiddleware) getOutputGrpc() (*grpc.ClientConn, *proto.MiddlewareClient) {
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

		client := proto.NewMiddlewareClient(conn)
		return conn, &client
	}

	return nil, nil
}
