package ack

import (
	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"k8s.io/klog"
	"net"
	"sync"
)

// @TODO Tests missing
type Manager struct {
	InputChannel  <-chan messages.Ack
	OutputChannel chan<- messages.Ack
	BrokerChannel chan messages.Ack
	BrokerAddress string
}

func (t *Manager) StartServer() {
	lis, err := net.Listen("tcp", t.BrokerAddress)
	if err != nil {
		klog.Errorf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
	proto.RegisterAckEventServer(grpcServer, &Server{
		clients:      make(map[string]proto.AckEvent_GetMessagesServer),
		InputChannel: t.BrokerChannel,
		mu:           sync.RWMutex{},
	})

	klog.V(1).Info("ACK-Manager listening...")
	err = grpcServer.Serve(lis)
	if err != nil {
		klog.Error(err)
	}
}

func (t *Manager) Start() {

	if t.ShouldStartAckBroker() {
		go t.StartServer()
	}

	for message := range t.InputChannel {
		t.OutputChannel <- message
		t.BrokerChannel <- message
	}
}

func (t *Manager) ShouldStartAckBroker() bool {
	if t.BrokerAddress == "" {
		klog.V(1).Info("ACK-Manager not available")
		return false
	}
	return true
}

func NewAckManager(inputChannel chan messages.Ack, outputChannel chan messages.Ack, brokerChannel chan messages.Ack) *Manager {
	return &Manager{
		InputChannel:  inputChannel,
		OutputChannel: outputChannel,
		BrokerChannel: brokerChannel,
		BrokerAddress: helpers.GetEnv("ACK_BROKER_ADDRESS", ""),
	}
}
