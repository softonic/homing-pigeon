package writers

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type writeAdapterMock struct {
	mock.Mock
	maxFlushSize int
	timeout      time.Duration
}

func (r *writeAdapterMock) ProcessMessages(msgs []*messages.Message) []*messages.Ack {
	acks := make([]*messages.Ack, 0)
	for _, msg := range msgs {
		ack, err := msg.Ack()
		if err == nil {
			acks = append(acks, ack)
		}
	}
	return acks}

func (r *writeAdapterMock) ShouldProcess(msgs []*messages.Message) bool {
	return len(msgs) >= r.maxFlushSize
}

func (r *writeAdapterMock) GetTimeout() time.Duration {
	return r.timeout
}

func TestAdapterProcessSingleMessage(t *testing.T) {
	writeAdapter := new(writeAdapterMock)
	writeAdapter.maxFlushSize = 1
	writeAdapter.timeout = time.Hour
	msgChannel := make(chan messages.Message, 1)
	ackChannel := make(chan messages.Ack, 1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}
	msgChannel <- messages.Message{
		Id:   0,
		Body: []byte("hola"),
	}

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, 1) && assert.Empty(t, msgChannel)
		},
		time.Millisecond*10,
		time.Millisecond,
	)
}

func TestAdapterProcessBulkMessages(t *testing.T) {
	writeAdapter := new(writeAdapterMock)
	writeAdapter.maxFlushSize = 3
	writeAdapter.timeout = time.Hour

	msgChannel := make(chan messages.Message, 4)
	ackChannel := make(chan messages.Ack, 4)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}

	for i := 0; i < writeAdapter.maxFlushSize; i++ {
		msgChannel <- messages.Message{
			Id:   uint64(i),
			Body: []byte("hola"),
		}

	}

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, writeAdapter.maxFlushSize) && assert.Empty(t, msgChannel)
		},
		time.Millisecond*10,
		time.Millisecond,
	)
}

func TestAdapterTimeoutProcessMessages(t *testing.T) {
	writeAdapter := new(writeAdapterMock)
	writeAdapter.maxFlushSize = 3
	writeAdapter.timeout = 0

	itemsToProcess := writeAdapter.maxFlushSize - 1
	msgChannel := make(chan messages.Message, 4)
	ackChannel := make(chan messages.Ack, 4)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}

	for i := 0; i < itemsToProcess; i++ {
		msgChannel <- messages.Message{
			Id:   uint64(i),
			Body: []byte("hola"),
		}
	}

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, itemsToProcess)
		},
		time.Millisecond*10,
		time.Millisecond,
	)
}
