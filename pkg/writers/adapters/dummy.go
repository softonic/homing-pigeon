package adapters

import (
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
)

type Dummy struct{}

func (d *Dummy) ProcessMessages(msgs *[]messages.Message) {
	for i := range *msgs {
		(*msgs)[i].Ack()
	}
}

func (d *Dummy) ShouldProcess(msgs []messages.Message) bool {
	return len(msgs) > 0
}

func (d *Dummy) GetTimeout() time.Duration {
	return time.Second
}
