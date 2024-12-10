package writers

import (
	"testing"
	"time"

	"github.com/softonic/homing-pigeon/mocks"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdapterProcessSingleMessage(t *testing.T) {

	writeAdapter := new(mocks.WriteAdapter)
	writeAdapter.On("GetTimeout").Return(time.Hour)
	writeAdapter.On("ShouldProcess", mock.Anything).Return(true)

	mockProcessMessages(writeAdapter)

	msgChannel := make(chan messages.Message, 1)
	ackChannel := make(chan messages.Message, 1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   msgChannel,
		AckChannel:   ackChannel,
	}
	msgChannel <- messages.Message{
		Id:   0,
		Body: []byte("hello"),
	}

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return len(ackChannel) == 1 && len(msgChannel) == 0
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
	writeAdapter.On("ShouldProcess", mock.MatchedBy(func(msgs []messages.Message) bool {
		return len(msgs) != expectedMessages
	})).Return(false)
	writeAdapter.On("ShouldProcess", mock.MatchedBy(func(msgs []messages.Message) bool {
		return len(msgs) == expectedMessages
	})).Return(true)

	mockProcessMessages(writeAdapter)

	msgChannel := make(chan messages.Message, expectedMessages+1)
	ackChannel := make(chan messages.Message, expectedMessages+1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   msgChannel,
		AckChannel:   ackChannel,
	}

	writeMessagesToChannel(&msgChannel)

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return len(ackChannel) == expectedMessages && len(msgChannel) == 0
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

	mockProcessMessages(writeAdapter)

	msgChannel := make(chan messages.Message, expectedMessages+1)
	ackChannel := make(chan messages.Message, expectedMessages+1)

	writer := Writer{
		WriteAdapter: writeAdapter,
		MsgChannel:   msgChannel,
		AckChannel:   ackChannel,
	}

	writeMessagesToChannel(&msgChannel)

	go writer.Start()

	assert.Eventually(
		t,
		func() bool {
			return expectedMessages == len(ackChannel) && len(msgChannel) == 0
		},
		time.Millisecond*500,
		time.Millisecond,
	)

	writeAdapter.AssertExpectations(t)
}

func writeMessagesToChannel(msgChannel *chan messages.Message) {
	*msgChannel <- messages.Message{
		Id:   uint64(1),
		Body: []byte("hello"),
	}
	*msgChannel <- messages.Message{
		Id:   uint64(2),
		Body: []byte("hello"),
	}
	*msgChannel <- messages.Message{
		Id:   uint64(3),
		Body: []byte("hello"),
	}
}

func getAcks(n int) []messages.Message {
	acks := make([]messages.Message, n)
	for i := 0; i < n; i++ {
		acks[i] = messages.Message{
			Id:   uint64(i),
			Body: []byte{0},
		}
	}
	return acks
}

func mockProcessMessages(writeAdapter *mocks.WriteAdapter) {
	writeAdapter.On("ProcessMessages", mock.MatchedBy(func(msgs *[]messages.Message) bool {
		return len(*msgs) == 3
	})).Maybe().Return(getAcks(3))

	writeAdapter.On("ProcessMessages", mock.MatchedBy(func(msgs *[]messages.Message) bool {
		return len(*msgs) == 1
	})).Maybe().Return(getAcks(1))
	writeAdapter.On("ProcessMessages", mock.MatchedBy(func(msgs *[]messages.Message) bool {
		return len(*msgs) == 2
	})).Maybe().Return(getAcks(2))
	writeAdapter.On("ProcessMessages", mock.MatchedBy(func(msgs *[]messages.Message) bool {
		return len(*msgs) == 0
	})).Maybe().Return(getAcks(0))
}
