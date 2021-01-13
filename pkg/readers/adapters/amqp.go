package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	amqpAdapter "github.com/softonic/homing-pigeon/pkg/readers/adapters/amqp"
	"github.com/streadway/amqp"
	"k8s.io/klog"
	"log"
	"os"
)

type Amqp struct {
	ConsumedMessages <-chan amqp.Delivery
	Conn             amqpAdapter.Connection
	Ch               amqpAdapter.Channel
	Notify           chan *amqp.Error
}

// @TODO detected race condition with closed channel
func (a *Amqp) Listen(msgChannel chan<- messages.Message) {
	defer a.Conn.Close()
	defer a.Ch.Close()

	go a.processMessages(msgChannel)
	klog.V(0).Infof(" [*] Waiting for messages. To exit press CTRL+C")
	select {}
}

func (a *Amqp) processMessages(writeChannel chan<- messages.Message) {
	for {
		select {
		case err := <-a.Notify:
			if err != nil {
				log.Fatalf("Error in connection: %s", err)
			}
			log.Println("Closed connection.")
			os.Exit(0)
			break
		case d := <-a.ConsumedMessages:
			msg := messages.Message{}
			msg.Id = d.DeliveryTag
			msg.Body = d.Body

			writeChannel <- msg
		}
	}
}

func (a *Amqp) HandleAck(ackChannel <-chan messages.Ack) {
	for ack := range ackChannel {
		if ack.Ack {
			err := a.Ch.Ack(ack.Id.(uint64), false)
			if err != nil {
				klog.Error(err)
			}
			continue
		}

		err := a.Ch.Nack(ack.Id.(uint64), false, false)
		if err != nil {
			klog.Error(err)
		}
	}
}
