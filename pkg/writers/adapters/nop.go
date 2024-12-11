package adapters

import (
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
)

type Nop struct{}

func (wa *Nop) ProcessMessages(msgs *[]messages.Message) {
	for i := range *msgs {
		(*msgs)[i].Ack()
	}
}

func (wa *Nop) GetTimeout() time.Duration {
	return time.Millisecond * 1000
}

func (wa *Nop) ShouldProcess(msgs []messages.Message) bool {
	return true
}
