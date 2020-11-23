package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"time"
)

type WriteAdapter interface {
	ProcessMessages(msgs []messages.Message) []messages.Ack
	ShouldProcess(msgs []messages.Message) bool
	GetTimeout() time.Duration
}
