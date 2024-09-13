package adapters

import (
	"fmt"
	"strconv"

	"github.com/softonic/homing-pigeon/pkg/messages"
)

type Dummy struct{}

func (d *Dummy) Listen(msgChannel chan<- messages.Message) {
	for i := 0; i < 100; i++ {
		msg := messages.Message{
			Id:   uint64(i),
			Body: []byte("my message " + strconv.Itoa(i)),
		}
		msgChannel <- msg
	}
}

func (d *Dummy) HandleAck(ackChannel <-chan messages.Message) {
	for ack := range ackChannel {
		fmt.Print("Acked " + strconv.Itoa(int(ack.Id.(uint64))) + "\n")
	}
}
