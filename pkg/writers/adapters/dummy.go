package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"time"
)

type Dummy struct{}

func (d *Dummy) ProcessMessages(msgs []messages.Message) []messages.Ack {
	var processedMsg []messages.Ack
	for i := 0; i < len(msgs); i++ {
		processedMsg = append(processedMsg, messages.Ack{
			Id:  uint64(i),
			Ack: true,
		})
	}
	return processedMsg
}

func (d *Dummy) ShouldProcess(msgs []messages.Message) bool {
	return len(msgs) > 0
}

func (d *Dummy) GetTimeout() time.Duration {
	return time.Second
}
