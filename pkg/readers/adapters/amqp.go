package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	amqpAdapter "github.com/softonic/homing-pigeon/pkg/readers/adapters/amqp"
	"github.com/streadway/amqp"
	"log"
)

type Amqp struct {
	ConsumedMessages <-chan amqp.Delivery
	Conn             amqpAdapter.Connection
	Ch               amqpAdapter.Channel
}

// @TODO detected race condition with closed channel
func (a *Amqp) Listen(msgChannel chan<- messages.Message) {
	defer a.Conn.Close()
	defer a.Ch.Close()

	go a.processMessages(msgChannel)
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {}
}

func (a *Amqp) processMessages(writeChannel chan<- messages.Message) {
	msg := messages.Message{}
	for d := range a.ConsumedMessages {
		msg.Id = d.DeliveryTag
		msg.Body = d.Body

		writeChannel <- msg
	}
}

func (a *Amqp) HandleAck(ackChannel <-chan messages.Ack) {
	for ack := range ackChannel {
		if ack.Ack {
			err := a.Ch.Ack(ack.Id.(uint64), false)
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

		err := a.Ch.Nack(ack.Id.(uint64), false, false)
		if err != nil {
			log.Fatal(err)
		}
	}
}
