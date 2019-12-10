package middleware

import (
	"testing"
)

func TestAdapterProcessSingleMessage(t *testing.T) {
	//writeAdapter := new(mocks.WriteAdapter)
	//writeAdapter.On("GetTimeout").Return(time.Hour)
	//writeAdapter.On("ShouldProcess", mock.Anything).Return(true)
	//
	//mockProcessMessages(writeAdapter)
	//
	//msgChannel := make(chan messages.Message, 1)
	//ackChannel := make(chan messages.Ack, 1)
	//
	//writer := Writer{
	//	WriteAdapter: writeAdapter,
	//	MsgChannel:   msgChannel,
	//	AckChannel:   ackChannel,
	//}
	//msgChannel <- messages.Message{
	//	Id:   0,
	//	Body: []byte("hello"),
	//}
	//
	//go writer.Start()
	//
	//assert.Eventually(
	//	t,
	//	func() bool {
	//		return len(ackChannel) == 1 && len(msgChannel) == 0
	//	},
	//	time.Millisecond*10,
	//	time.Millisecond,
	//)
	//writeAdapter.AssertExpectations(t)
}