package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
)

type Nop struct {}

func (wa *Nop) ProcessMessages(msgs []messages.Message) []messages.Ack {
	acks := make([]messages.Ack, 0)
	for _, msg := range msgs {
		acks = append(acks, msg.Ack())
	}
	return acks
}

func (wa *Nop) GetTimeoutInMs() int64 {
	return int64(1000)
}

func (wa *Nop) ShouldProcess(msgs []messages.Message) bool {
	return true
}