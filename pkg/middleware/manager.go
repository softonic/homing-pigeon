package middleware

import (
	"context"
	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// @TODO Tests missing
type MiddlwareManager struct {
	MsgIncomingChannel <-chan messages.Message
	MsgOutgoingChannel chan<- messages.Message
	AckIncomingChannel <-chan messages.Ack
	AckOutgoingChannel chan<- messages.Ack
	MiddlewareAddress  string
	BrokerPort         string
}

func (m *MiddlwareManager) HandleInput() {
	if m.isMiddlewareNotAvailable() {
		klog.V(1).Infof("Middlewares not available")
		for message := range m.MsgIncomingChannel {
			m.MsgOutgoingChannel <- message
		}
	}

	klog.V(1).Infof("Middlewares available")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(m.MiddlewareAddress, opts...)
	if err != nil {
		klog.Errorf("fail to dial: %v", err)
	}

	klog.V(1).Infof("Middlewares connected")

	defer conn.Close()
	client := proto.NewMiddlewareClient(conn)

	for message := range m.MsgIncomingChannel {
		klog.V(5).Infof("Sending message to proto")
		start := time.Now()

		data, err := client.Handle(context.Background(), &proto.Data{
			Body: message.Body,
		})
		if err != nil {
			klog.Errorf("What happens!? %v", err)
		}

		elapsed := time.Since(start)
		klog.V(5).Infof("Middlewares took %s", elapsed)

		message.Body = data.GetBody()

		m.MsgOutgoingChannel <- message
	}
}

func (m *MiddlwareManager) HandleOutput() {
	if m.isBrokerNotAvailable() {
		for message := range m.AckIncomingChannel {
			m.AckOutgoingChannel <- message
		}
	}

	brokerChannel := getBrokerChannel()

	// Start Broker server
	go func() {
		lis, err := net.Listen("tcp", m.BrokerPort)
		if err != nil {
			klog.Errorf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(10))
		proto.RegisterBrokerServer(grpcServer, &Broker{
			clients:      make(map[string]proto.Broker_GetMessageServer),
			InputChannel: brokerChannel,
			mu:           sync.RWMutex{},
		})

		klog.V(1).Info("Response Broker listening...")
		err = grpcServer.Serve(lis)
		if err != nil {
			klog.Error(err)
		}
	}()

	for message := range m.AckIncomingChannel {
		m.AckOutgoingChannel <- message
		brokerChannel <- message
	}
}

func (m *MiddlwareManager) isMiddlewareNotAvailable() bool {
	return m.MiddlewareAddress == ""
}

func (m *MiddlwareManager) isBrokerNotAvailable() bool {
	return m.BrokerPort == ""
}

func getBrokerChannel() chan messages.Ack {
	bufLen, err := strconv.Atoi(os.Getenv("ACK_BUFFER_LENGTH"))
	if err != nil {
		bufLen = 0
	}

	return make(chan messages.Ack, bufLen)
}

func NewMiddlewareManager(msgCh1 chan messages.Message, msgCh2 chan messages.Message, ackCh1 chan messages.Ack, ackCh2 chan messages.Ack) *MiddlwareManager {
	return &MiddlwareManager{
		MsgIncomingChannel: msgCh1,
		MsgOutgoingChannel: msgCh2,
		AckIncomingChannel: ackCh1,
		AckOutgoingChannel: ackCh2,
		MiddlewareAddress:  helpers.GetEnv("MIDDLEWARES_SOCKET", ""),
		BrokerPort:         helpers.GetEnv("RESPONSE_BROKER_PORT", ""),
	}
}
