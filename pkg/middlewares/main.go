package main

import (
	"context"
	pb "github.com/softonic/homing-pigeon/middleware"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"log"
	"net"
)

type MiddlewareServer struct {
	client *pb.MiddlewareClient
	pb.UnimplementedMiddlewareServer
}

func (m *MiddlewareServer) Handle(ctx context.Context, req *pb.Data) (*pb.Data, error) {

	//log.Printf("Pre-Processing %v", *req)
	// Do things
	//---------------------

	resp := req
	if m.client != nil {
		var err error
		resp, err = (*m.client).Handle(context.Background(), req)
		if err != nil {
			log.Fatalf("What happens!? %v", err)
		}
	}

	//log.Printf("Post-Processing %v", *resp)
	// Do things

	return resp, nil
}

func main() {
	socket := getEnv("IN_SOCKET", "")

	err := os.RemoveAll(socket)
	if err != nil {
		log.Fatalf("Failed to remove socket: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(socket), 0775)
	if err != nil {
		log.Fatalf("Error creating socket directory: %v", err)
	}

	lis, err := net.Listen("unix", socket)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	conn, client := getGrpc()
	defer conn.Close()

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	pb.RegisterMiddlewareServer(grpcServer, &MiddlewareServer{
		client: client,
	})

	log.Print("Start listening")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

func getGrpc() (*grpc.ClientConn, *pb.MiddlewareClient){
	nextSocketAddr := getEnv("OUT_SOCKET", "")
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

func getEnv(envVarName string, defaultValue string) string {
	value := os.Getenv(envVarName)
	if value != "" {
		return value
	}
	return defaultValue
}
