package adapters

import "github.com/softonic/homing-pigeon/pkg/messages"

// ReadAdapter is an interface for specific reader implementations.
type ReadAdapter interface {
	Listen(msgChannel chan<- messages.Message)
	HandleAck(ackChannel <-chan messages.Message)
}
