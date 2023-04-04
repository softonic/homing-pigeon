package middleware

import (
	"github.com/google/uuid"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"k8s.io/klog"
	"sync"
)

type Broker struct {
	clients      map[string]proto.Broker_GetMessageServer
	InputChannel <-chan messages.Ack
	mu           sync.RWMutex
}

func (b *Broker) AddClient(uid string, client proto.Broker_GetMessageServer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[uid] = client
	klog.V(1).Info("New client connected")
}

func (b *Broker) GetClientsCopy() []proto.Broker_GetMessageServer {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var clients []proto.Broker_GetMessageServer
	for _, client := range b.clients {
		clients = append(clients, client)
	}
	return clients
}

func (b *Broker) GetMessage(req *proto.EmptyRequest, client proto.Broker_GetMessageServer) error {
	uid := uuid.Must(uuid.NewRandom()).String()

	b.AddClient(uid, client)

	for message := range b.InputChannel {
		klog.V(1).Info("Sending ACK to clients")
		for _, client := range b.GetClientsCopy() {
			err := client.Send(&proto.Message{
				Body:    message.Body,
				Success: message.Ack,
			})

			if err != nil {
				klog.Errorf("Error sending to grpc client: %v", err)
			}
		}
	}
	return nil
}
