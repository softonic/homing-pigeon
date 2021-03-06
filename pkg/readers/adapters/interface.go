package adapters

import "github.com/softonic/homing-pigeon/pkg/messages"

type ReadAdapter interface {
	Listen(msgChannel chan<- messages.Message)
	HandleAck(ackChannel <-chan messages.Ack)
}
