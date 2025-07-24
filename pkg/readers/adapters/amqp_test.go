package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/softonic/homing-pigeon/mocks"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestProcessMessage(t *testing.T) {
	expectedMessages := 1
	msgChannel := make(chan messages.Message, expectedMessages+1)
	consumedMessages := make(chan amqp.Delivery, expectedMessages+1)

	obj := Amqp{
		ConsumedMessages: consumedMessages,
		Conn:             new(mocks.Connection),
		Ch:               new(mocks.Channel),
	}

	consumedMessages <- amqp.Delivery{
		DeliveryTag: 42,
		Body:        []byte("Hello!"),
	}
	go obj.Listen(context.Background(), msgChannel)

	assert.Eventually(
		t,
		func() bool {
			return assert.Len(t, msgChannel, expectedMessages)
		},
		time.Millisecond*500,
		time.Millisecond,
	)

	msg := <-msgChannel
	assert.Equal(t, uint64(42), msg.Id)
	assert.Equal(t, []byte("Hello!"), msg.Body)
}

func TestHandleAck(t *testing.T) {
	expectedMessages := 1
	ackChannel := make(chan messages.Message, expectedMessages+1)

	channel := new(mocks.Channel)
	expectedId := uint64(42)
	channel.On("Ack", expectedId, false).Once().Return(nil)

	obj := Amqp{
		ConsumedMessages: nil,
		Conn:             nil,
		Ch:               channel,
	}

	msg := messages.Message{
		Id:   expectedId,
		Body: []byte("Hello!"),
	}
	msg.Ack()

	ackChannel <- msg

	go obj.HandleAck(ackChannel)

	assert.Eventually(
		t,
		func() bool {
			return channel.AssertExpectations(t) && channel.AssertNotCalled(t, "Nack")
		},
		time.Millisecond*10,
		time.Millisecond,
	)
}

func TestHandleNack(t *testing.T) {
	expectedMessages := 1
	ackChannel := make(chan messages.Message, expectedMessages+1)

	channel := new(mocks.Channel)
	expectedId := uint64(42)
	channel.On("Nack", expectedId, false, false).Once().Return(nil)

	obj := Amqp{
		ConsumedMessages: nil,
		Conn:             nil,
		Ch:               channel,
	}

	msg := messages.Message{
		Id:   expectedId,
		Body: []byte("Hello!"),
	}
	msg.Nack()

	ackChannel <- msg

	go obj.HandleAck(ackChannel)

	assert.Eventually(
		t,
		func() bool {
			return channel.AssertExpectations(t) && channel.AssertNotCalled(t, "Ack")
		},
		time.Millisecond*10,
		time.Millisecond,
	)
}

func TestHandleMixedAcks(t *testing.T) {
	expectedMessages := 1
	ackChannel := make(chan messages.Message, expectedMessages+1)

	channel := new(mocks.Channel)
	expectedAckId := uint64(42)
	channel.On("Ack", expectedAckId, false).Once().Return(nil)
	expectedNackId := uint64(50)
	channel.On("Nack", expectedNackId, false, false).Once().Return(nil)

	obj := Amqp{
		ConsumedMessages: nil,
		Conn:             nil,
		Ch:               channel,
	}

	msgAck := messages.Message{
		Id:   expectedAckId,
		Body: []byte("Hello!"),
	}
	msgAck.Ack()

	msgNack := messages.Message{
		Id:   expectedNackId,
		Body: []byte("Hello!"),
	}
	msgNack.Nack()

	ackChannel <- msgAck
	ackChannel <- msgNack

	go obj.HandleAck(ackChannel)

	assert.Eventually(
		t,
		func() bool {
			return channel.AssertExpectations(t)
		},
		time.Millisecond*10,
		time.Millisecond,
	)
}
