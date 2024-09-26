package adapters

import (
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
)

type WriteAdapter interface {
	ProcessMessages(msgs []messages.Message) []messages.Message
	ShouldProcess(msgs []messages.Message) bool
	GetTimeout() time.Duration
}
