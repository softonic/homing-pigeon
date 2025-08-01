package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
)

func TestProduceMessageQuantity(t *testing.T) {
	expectedMessages := 100
	msgChannel := make(chan messages.Message, expectedMessages+1)
	obj := new(Dummy)
	obj.Listen(context.Background(), msgChannel)
	assert.Len(t, msgChannel, expectedMessages)
}

func TestAcksAreRead(t *testing.T) {
	ackChannel := make(chan messages.Message, 2)

	msg := messages.Message{
		Id:   uint64(1),
		Body: []byte("Hello!"),
	}
	msg.Ack()

	ackChannel <- msg

	obj := new(Dummy)
	go obj.HandleAck(context.Background(), ackChannel)

	assert.Eventually(t, func() bool {
		return assert.Empty(t, ackChannel)
	}, time.Millisecond*10, time.Millisecond)
}
