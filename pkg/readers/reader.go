package readers

import (
	"context"

	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/readers/adapters"
	"k8s.io/klog"
)

// Reader represents a reader for handling messages.
type Reader struct {
	ReadAdapter adapters.ReadAdapter
	MsgChannel  chan<- messages.Message
	AckChannel  <-chan messages.Message
}

// Start starts the reader.
func (r *Reader) Start(ctx context.Context) {
	go r.ReadAdapter.HandleAck(ctx, r.AckChannel)
	r.ReadAdapter.Listen(ctx, r.MsgChannel)
}

// NewReader creates a new Reader instance.
func NewReader(inputChannel chan messages.Message, ackChannel chan messages.Message) (*Reader, error) {

	var err error
	var readAdapter adapters.ReadAdapter
	adapter := helpers.GetEnv("READ_ADAPTER", "AMQP")

	switch adapter {
	case "AMQP":
		readAdapter, err = adapters.NewAMQPAdapter()
	default:
		klog.Warning("Reader not defined, using dummy implementation")
		readAdapter, err = &adapters.Dummy{}, nil
	}

	return &Reader{
		ReadAdapter: readAdapter,
		MsgChannel:  inputChannel,
		AckChannel:  ackChannel,
	}, err
}
