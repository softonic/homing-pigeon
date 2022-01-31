package ack

import (
	"github.com/google/uuid"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"k8s.io/klog"
	"sync"
)

type Server struct {
	clients      map[string]proto.AckEvent_GetMessagesServer
	InputChannel <-chan messages.Ack
	mu           sync.RWMutex
}

func (s *Server) AddClient(uid string, client proto.AckEvent_GetMessagesServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[uid] = client
	klog.V(1).Info("New client connected")
}

func (s *Server) GetClientsCopy() []proto.AckEvent_GetMessagesServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var clients []proto.AckEvent_GetMessagesServer
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

func (s *Server) GetMessages(req *proto.EmptyRequest, client proto.AckEvent_GetMessagesServer) error {
	uid := uuid.Must(uuid.NewRandom()).String()

	s.AddClient(uid, client)

	for message := range s.InputChannel {
		klog.V(1).Info("Sending ACK to clients")
		for _, client := range s.GetClientsCopy() {
			err := client.Send(&proto.Message{
				Body: message.Body,
				Ack:  message.Ack,
			})

			if err != nil {
				klog.Errorf("Error sending to grpc client: %v", err)
			}
		}
	}
	return nil
}
