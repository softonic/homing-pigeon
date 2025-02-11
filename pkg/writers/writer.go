package writers

import (
	"sync"
	"time"

	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/writers/adapters"
	"k8s.io/klog"
)

type Writer struct {
	MsgChannel   <-chan messages.Message
	AckChannel   chan<- messages.Message
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

func (ew *Writer) processAllMessages(msgChannel <-chan messages.Message, ackChannel chan<- messages.Message) {
	for msg := range msgChannel {
		ew.appendMessage(msg)
		if ew.shouldProcess() {
			go ew.trigger(ackChannel)
		}
	}
}

func (ew *Writer) timeout(ackChannel chan<- messages.Message) {
	for {
		time.Sleep(ew.WriteAdapter.GetTimeout())
		go ew.trigger(ackChannel)
	}
}

func (ew *Writer) trigger(ackChannel chan<- messages.Message) {
	ew.mutex.Lock()

	// filter out messages explicitly n-acked
	writeableMessages := make([]messages.Message, 0)
	for _, msg := range ew.msgs {
		if !msg.IsNacked() {
			writeableMessages = append(writeableMessages, msg)
		}
	}
	ew.WriteAdapter.ProcessMessages(&writeableMessages)
	ew.sendAcks(ew.msgs, ackChannel)
	ew.msgs = make([]messages.Message, 0)

	ew.mutex.Unlock()
}

func (ew *Writer) sendAcks(acks []messages.Message, ackChannel chan<- messages.Message) {
	for _, ack := range acks {
		ackChannel <- ack
	}
}

func NewWriter(outputChannel chan messages.Message, ackChannel chan messages.Message) (*Writer, error) {

	var err error
	var writeAdapter adapters.WriteAdapter
	adapter := helpers.GetEnv("WRITE_ADAPTER", "ELASTIC")

	switch adapter {
	case "ELASTIC":
		writeAdapter, err = adapters.NewElasticsearchAdapter()
	default:
		klog.Warning("Writer not defined, using dummy implementation")
		writeAdapter, err = &adapters.Dummy{}, nil
	}

	return &Writer{
		MsgChannel:   outputChannel,
		AckChannel:   ackChannel,
		WriteAdapter: writeAdapter,
	}, err
}
