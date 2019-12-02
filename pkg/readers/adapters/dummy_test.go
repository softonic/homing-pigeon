package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProduceMessageQuantity(t *testing.T) {
	expectedMessages := 100
	msgChannel := make(chan messages.Message, expectedMessages+1)

	obj := new(Dummy)
	obj.Listen(msgChannel)

	assert.Len(t, msgChannel, expectedMessages)
}

func TestAcksAreRead(t *testing.T) {
	ackChannel := make(chan messages.Ack, 2)
	ackChannel <- messages.Ack{
		Id: uint64(1),
		Ack: true,
	}

	obj := new(Dummy)
	go obj.HandleAck(ackChannel)

	assert.Eventually(t, func() bool {
		return assert.Empty(t, ackChannel)
	}, time.Millisecond*10, time.Millisecond)
}