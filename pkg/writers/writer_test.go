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
	NumMsgsToProcess int
}

func (r *writeAdapterMock) ProcessMessages(msgs []*messages.Message) []*messages.Ack {
	return []*messages.Ack{
		{
			Id:  0,
			Ack: true,
		},
	}
}

func (r *writeAdapterMock) ShouldProcess(msgs []*messages.Message) bool {
		return len(msgs) >= r.NumMsgsToProcess
}

func (r *writeAdapterMock) GetTimeout() time.Duration {
		return time.Hour
}

func TestAdapterProcessSingleMessage(t *testing.T) {
	writeAdapter := new(writeAdapterMock)
	writeAdapter.NumMsgsToProcess = 1
	msgChannel := make(chan messages.Message, 1)
	ackChannel := make(chan messages.Ack, 1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}

	msgChannel <- messages.Message{
		Id:   1,
		Body: []byte("hola"),
	}

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, 1)
		},
		time.Millisecond * 10,
		time.Millisecond,
	)
}

func TestAdapterProcessBulkMessages(t *testing.T) {
	writeAdapter := new(writeAdapterMock)
	writeAdapter.NumMsgsToProcess = 3
	msgChannel := make(chan messages.Message, 1)
	ackChannel := make(chan messages.Ack, 1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}

	for i := 1; i <= writeAdapter.NumMsgsToProcess; i++ {
		msgChannel <- messages.Message{
			Id:   uint64(i),
			Body: []byte("hola"),
		}
	}

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, 1)
		},
		time.Millisecond * 10,
		time.Millisecond,
	)
}
