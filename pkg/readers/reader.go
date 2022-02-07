package readers

import (
	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/readers/adapters"
	"k8s.io/klog"
)

type Reader struct {
	ReadAdapter adapters.ReadAdapter
	MsgChannel  chan<- messages.Message
	AckChannel  <-chan messages.Ack
}

func (r *Reader) Start() {
	go r.ReadAdapter.HandleAck(r.AckChannel)
	r.ReadAdapter.Listen(r.MsgChannel)
}

func NewReader(inputChannel chan messages.Message, ackChannel chan messages.Ack) (*Reader, error) {

	var err error
	var readAdapter adapters.ReadAdapter
	adapter := helpers.GetEnv("READ_ADAPTER", "AMQP")

	switch adapter {
	case "AMQP":
		readAdapter, err = NewAMQPAdapter()
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

func NewAMQPAdapter() (adapters.ReadAdapter, error) {

	amqpConfig, err := adapters.NewAmqpConfig()
	if err != nil {
		return nil, err
	}
	return adapters.NewAmqpReaderAdapter(amqpConfig)
}
