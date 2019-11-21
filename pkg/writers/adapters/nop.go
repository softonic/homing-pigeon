package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"time"
)

type Nop struct {}

func (wa *Nop) ProcessMessages(msgs []*messages.Message) []*messages.Ack {
	acks := make([]*messages.Ack, 0)
	for _, msg := range msgs {
		ack, err := msg.Ack()
		if err == nil {
			acks = append(acks, ack)
		}
	}
	return acks
}

func (wa *Nop) GetTimeout() time.Duration {
	return time.Millisecond * 1000
}

func (wa *Nop) ShouldProcess(msgs []*messages.Message) bool {
	return true
}