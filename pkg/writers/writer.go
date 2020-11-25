package writers

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/writers/adapters"
	"sync"
	"time"
)

type Writer struct {
	MsgChannel   <-chan messages.Message
	AckChannel   chan<- messages.Ack
	msgs         []messages.Message
	WriteAdapter adapters.WriteAdapter
	mutex        *sync.Mutex
}

func (ew *Writer) Start() {
	ew.mutex = &sync.Mutex{}
	go ew.timeout(ew.AckChannel)
	ew.processAllMessages(ew.MsgChannel, ew.AckChannel)
}

func (ew *Writer) appendMessage(msg messages.Message) {
	ew.mutex.Lock()
	ew.msgs = append(ew.msgs, msg)
	ew.mutex.Unlock()
}
func (ew *Writer) shouldProcess() bool {
	ew.mutex.Lock()
	res := ew.WriteAdapter.ShouldProcess(ew.msgs)
	ew.mutex.Unlock()
	return res
}

func (ew *Writer) processAllMessages(msgChannel <-chan messages.Message, ackChannel chan<- messages.Ack) {
	for msg := range msgChannel {
		ew.appendMessage(msg)
		if ew.shouldProcess() {
			go ew.trigger(ackChannel)
		}
	}
}

func (ew *Writer) timeout(ackChannel chan<- messages.Ack) {
	for {
		time.Sleep(ew.WriteAdapter.GetTimeout())
		go ew.trigger(ackChannel)
	}
}

func (ew *Writer) trigger(ackChannel chan<- messages.Ack) {
	ew.mutex.Lock()

	acks := ew.WriteAdapter.ProcessMessages(ew.msgs)
	ew.msgs = make([]messages.Message, 0)
	ew.sendAcks(acks, ackChannel)

	ew.mutex.Unlock()
}

func (ew *Writer) sendAcks(acks []messages.Ack, ackChannel chan<- messages.Ack) {
	for _, ack := range acks {
		ackChannel <- ack
	}
}
