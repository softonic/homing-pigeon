package adapters

import "github.com/softonic/homing-pigeon/pkg/messages"

type WriteAdapter interface {
	ProcessMessages(msgs []messages.Message) []messages.Ack
	ShouldProcess(msgs []messages.Message) bool
	GetTimeoutInMs() int64
}