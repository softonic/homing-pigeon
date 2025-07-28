package middleware

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"time"

	. "github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog"
)

// @TODO Tests missing
type UnimplementedMiddleware struct {
	client *proto.MiddlewareClient
	proto.UnimplementedMiddlewareServer
}

func (b *UnimplementedMiddleware) Next(req *proto.Data) (*proto.Data, error) {
	if b.client != nil {
		klog.V(0).Info("processing next middleware")
		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 31*time.Second)
		nextResp, err := (*b.client).Handle(ctxTimeout, req, grpc.WaitForReady(true))
		cancelTimeout()
		if err != nil {
			klog.Errorf("Next middleware error %v", err)
			return nil, err
		} else {
			klog.V(0).Info("next middleware processed")
			return nextResp, nil
		}
	}
	// if there is no next middleware return the same request
	return req, nil
}

func (b *UnimplementedMiddleware) Listen(middleware proto.MiddlewareServer) {
	lis, err := net.Listen("unix", b.getInputSocket())
	if err != nil {
		klog.Errorf("Failed to listen: %v", err)
	}

	conn, client := b.getOutputGrpc()
	defer conn.Close()

	b.client = client

	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(10),
		grpc.MaxRecvMsgSize(defaultMaxMessageSize),
		grpc.MaxSendMsgSize(defaultMaxMessageSize),
	)
	proto.RegisterMiddlewareServer(grpcServer, middleware)

	klog.V(0).Info("Start listening")
	err = grpcServer.Serve(lis)
	if err != nil {
		klog.Error(err)
	}
}

func (b *UnimplementedMiddleware) getInputSocket() string {
	socket := GetEnv("IN_SOCKET", "")
	if _, err := os.Stat(socket); err == nil {
		klog.Infof("Removing existing socket %s", socket)
		err := os.Remove(socket)
		if err != nil {
			klog.Errorf("Failed to remove socket: %v", err)
		}
	} else {
		err = os.MkdirAll(filepath.Dir(socket), 0775)
		if err != nil {
			klog.Errorf("Error creating socket directory: %v", err)
		}
	}
	return socket
}

func (b *UnimplementedMiddleware) getOutputGrpc() (*grpc.ClientConn, *proto.MiddlewareClient) {
	nextSocketAddr := GetEnv("OUT_SOCKET", "")
	if nextSocketAddr == "" {
		return nil, nil
	}
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(defaultRetryPolicy))

	conn, err := grpc.NewClient(nextSocketAddr, opts...)
	if err != nil {
		klog.Errorf("failed to create OUT_SOCKET client: %v", err)
		return nil, nil
	}

	klog.V(0).Info("Connected to the next middleware")

	client := proto.NewMiddlewareClient(conn)
	return conn, &client
}
