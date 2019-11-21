package writers

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/writers/adapters"
	"time"
)

type Writer struct {
	MsgChannel   *chan messages.Message
	AckChannel   *chan messages.Ack
	msgs         []*messages.Message
	WriteAdapter adapters.WriteAdapter
}

func (ew *Writer) Start() {
	go ew.timeout()

	for msg := range *ew.MsgChannel {
		msg := msg
		ew.msgs = append(ew.msgs, &msg)

		if ew.WriteAdapter.ShouldProcess(ew.msgs) {
			go ew.trigger(ew.msgs)
			ew.msgs = make([]*messages.Message, 0)
		}
	}
}

func (ew *Writer) timeout() {
	for {
		time.Sleep(ew.WriteAdapter.GetTimeout())

		go ew.trigger(ew.msgs)
		ew.msgs = make([]*messages.Message, 0)
	}
}

func (ew *Writer) trigger(msgs []*messages.Message) {
	acks := ew.WriteAdapter.ProcessMessages(msgs)
	ew.sendAcks(acks)
}

func (ew *Writer) sendAcks(acks []*messages.Ack) {
	for _, ack := range acks {
		*ew.AckChannel <- *ack
	}
}
