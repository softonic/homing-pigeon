package writers

import (
	"github.com/softonic/homing-pigeon/pkg/writers/adapters"
	"time"
)

type Writer struct {
	WriteChannel *chan string
	AckChannel   *chan string
	bulkMsgs     []string
	WriteAdapter adapters.WriteAdapter
}

func (ew *Writer) Start() {
	go ew.timeout()

	for msg := range *ew.WriteChannel {
		ew.bulkMsgs = append(ew.bulkMsgs, msg)

		if ew.WriteAdapter.ShouldProcess(ew.bulkMsgs) {
			go ew.trigger(ew.bulkMsgs)
			ew.bulkMsgs = make([]string, 0)
		}
	}
}

func (ew *Writer) timeout() {
	for {
		time.Sleep(10000 * time.Millisecond)

		go ew.trigger(ew.bulkMsgs)
		ew.bulkMsgs = make([]string, 0)
	}
}

func (ew *Writer) trigger(bulkMsgs []string) {
	acks := ew.WriteAdapter.ProcessMessages(bulkMsgs)
	ew.sendAcks(acks)
}

func (ew *Writer) sendAcks(acks []string) {
	for _, ack := range acks {
		*ew.AckChannel <- ack
	}
}
