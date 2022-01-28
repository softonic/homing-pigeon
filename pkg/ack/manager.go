package ack

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"net"
	"sync"
)

// @TODO Tests missing
type Manager struct {
	InputChannel   <-chan messages.Ack
	OutputChannel  chan<- messages.Ack
	ServiceChannel chan messages.Ack
	ServiceAddress string
}

func (t *Manager) StartServer() {
	lis, err := net.Listen("tcp", t.ServiceAddress)
	if err != nil {
		klog.Errorf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	proto.RegisterAckServiceServer(grpcServer, &Server{
		clients:      make(map[string]proto.AckService_GetMessagesServer),
		InputChannel: t.ServiceChannel,
		mu:           sync.RWMutex{},
	})

	klog.V(1).Info("ACK-Manager listening...")
	err = grpcServer.Serve(lis)
	if err != nil {
		klog.Error(err)
	}
}

func (t *Manager) Start() {

	if t.ShouldStartAckService() {
		go t.StartServer()
	}

	for message := range t.InputChannel {
		t.OutputChannel <- message
		t.ServiceChannel <- message
	}
}

func (t *Manager) ShouldStartAckService() bool {
	if t.ServiceAddress == "" {
		klog.V(1).Info("ACK-Manager not available")
		return false
	}
	return true
}
