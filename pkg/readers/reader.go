package readers

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/readers/adapters"
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

func NewAMQPReader(inputChannel chan messages.Message, ackChannel chan messages.Ack) (*Reader, error) {

	amqpConfig, err := adapters.NewAmqpConfig()
	if err != nil {
		return nil, err
	}
	adapter, err := adapters.NewAmqpReaderAdapter(amqpConfig)
	if err != nil {
		return nil, err
	}
	return &Reader{
		ReadAdapter: adapter,
		MsgChannel:  inputChannel,
		AckChannel:  ackChannel,
	}, nil
}
