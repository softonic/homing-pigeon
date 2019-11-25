package writers

import (
	"github.com/softonic/homing-pigeon/mocks"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestAdapterProcessSingleMessage(t *testing.T) {
	ack := []*messages.Ack{
		{
			Id:  0,
			Ack: false,
		},
	}

	writeAdapter := new(mocks.WriteAdapter)
	writeAdapter.On("GetTimeout").Return(time.Hour)
	writeAdapter.On("ShouldProcess", mock.Anything).Return(true)
	writeAdapter.On("ProcessMessages", mock.Anything).Return(ack)

	msgChannel := make(chan messages.Message, 1)
	ackChannel := make(chan messages.Ack, 1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}
	msgChannel <- messages.Message{
		Id:   0,
		Body: []byte("hello"),
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
	writeAdapter.AssertExpectations(t)
}

func TestAdapterProcessBulkMessages(t *testing.T) {
	expectedMessages := 3
	writeAdapter := new(mocks.WriteAdapter)
	writeAdapter.On("GetTimeout").Return(time.Hour)
	writeAdapter.On("ShouldProcess", mock.MatchedBy(func(msgs []*messages.Message) bool {
		return len(msgs) != expectedMessages
	})).Return(false)
	writeAdapter.On("ShouldProcess", mock.MatchedBy(func(msgs []*messages.Message) bool {
		return len(msgs) == expectedMessages
	})).Return(true)
	writeAdapter.On("ProcessMessages", mock.Anything).Return([]*messages.Ack{
		{
			Id:  0,
			Ack: false,
		},
		{
			Id:  0,
			Ack: false,
		},
		{
			Id:  0,
			Ack: false,
		},
	})
	msgChannel := make(chan messages.Message, expectedMessages+1)
	ackChannel := make(chan messages.Ack, expectedMessages+1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}

	writeMessagesToChannel(&msgChannel)

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, expectedMessages) && assert.Empty(t, msgChannel)
		},
		time.Millisecond*10,
		time.Millisecond,
	)

	writeAdapter.AssertExpectations(t)
}

func TestAdapterTimeoutProcessMessages(t *testing.T) {
	expectedMessages := 3
	writeAdapter := new(mocks.WriteAdapter)
	writeAdapter.On("GetTimeout").Return(time.Duration(0))
	writeAdapter.On("ShouldProcess", mock.Anything).Return(false)
	writeAdapter.On("ProcessMessages", mock.MatchedBy(func(msgs []*messages.Message) bool {
		return len(msgs) == expectedMessages
	})).Return([]*messages.Ack{
		{
			Id:  0,
			Ack: false,
		},
		{
			Id:  0,
			Ack: false,
		},
		{
			Id:  0,
			Ack: false,
		},
	})
	writeAdapter.On("ProcessMessages", mock.MatchedBy(func(msgs []*messages.Message) bool {
		return len(msgs) != expectedMessages
	})).Return([]*messages.Ack{})
	msgChannel := make(chan messages.Message, expectedMessages+1)
	ackChannel := make(chan messages.Ack, expectedMessages+1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   &msgChannel,
		AckChannel:   &ackChannel,
	}

	writeMessagesToChannel(&msgChannel)

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, ackChannel, expectedMessages) && assert.Empty(t, msgChannel)
		},
		time.Millisecond*10,
		time.Millisecond,
	)

	writeAdapter.AssertExpectations(t)
}

func writeMessagesToChannel(msgChannel *chan messages.Message) {
	*msgChannel <- messages.Message{
		Id:   uint64(0),
		Body: []byte("hello"),
	}
	*msgChannel <- messages.Message{
		Id:   uint64(1),
		Body: []byte("hello"),
	}
	*msgChannel <- messages.Message{
		Id:   uint64(2),
		Body: []byte("hello"),
	}
}
