package writers

import (
	"context"
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

func (ew *Writer) Start(ctx context.Context) {
	ew.mutex = &sync.Mutex{}
	ew.processAllMessages(ctx, ew.MsgChannel, ew.AckChannel)
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

func (ew *Writer) processAllMessages(ctx context.Context, msgChannel <-chan messages.Message, ackChannel chan<- messages.Message) {
	for {
		select {
		case <-ctx.Done():
			if ew.shouldProcess() {
				ew.trigger(ackChannel)
			}
			return
		case msg := <-msgChannel:
			ew.appendMessage(msg)
			if ew.shouldProcess() {
				go ew.trigger(ackChannel)
			}
		case <-time.After(ew.WriteAdapter.GetTimeout()):
			go ew.trigger(ackChannel)
		}
	}
}

func (ew *Writer) trigger(ackChannel chan<- messages.Message) {
	ew.mutex.Lock()

	ew.WriteAdapter.ProcessMessages(&ew.msgs)
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
