package readers

import (
	"sync"
	"testing"

	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/mock"
)

type readAdapterMock struct {
	mock.Mock
	wg sync.WaitGroup
}

func (r *readAdapterMock) Listen(msgChannel chan<- messages.Message) {
	r.Called(msgChannel)
}
func (r *readAdapterMock) HandleAck(ackChannel <-chan messages.Message) {
	r.Called(ackChannel)
	r.wg.Done()
}

func TestAdapterIsStarted(t *testing.T) {
	readAdapterMock := new(readAdapterMock)
	msgChannel := make(chan<- messages.Message)
	ackChannel := make(<-chan messages.Message)
	readAdapterMock.wg.Add(1)

	readAdapterMock.On("Listen", msgChannel)
	readAdapterMock.On("HandleAck", ackChannel)

	reader := Reader{
		ReadAdapter: readAdapterMock,
		MsgChannel:  msgChannel,
		AckChannel:  ackChannel,
	}
	reader.Start()
	readAdapterMock.wg.Wait()

	readAdapterMock.AssertExpectations(t)
}
