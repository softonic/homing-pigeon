package adapters

import (
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
)

type Dummy struct{}

func (d *Dummy) ProcessMessages(msgs []messages.Message) []messages.Message {
	var processedMsg []messages.Message
	for i := 0; i < len(msgs); i++ {
		processedMsg = append(processedMsg, messages.Message{
			Id:   uint64(i),
			Body: []byte{1},
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
