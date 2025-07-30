package adapters

import (
	"context"

	"github.com/softonic/homing-pigeon/pkg/messages"
)

// ReadAdapter is an interface for specific reader implementations.
type ReadAdapter interface {
	Listen(ctx context.Context, msgChannel chan<- messages.Message)
	HandleAck(ackChannel <-chan messages.Message)
}
