package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	amqpAdapter "github.com/softonic/homing-pigeon/pkg/readers/adapters/amqp"
	"github.com/streadway/amqp"
	"log"
)

type Amqp struct {
	Config amqpAdapter.Config
	conn   *amqp.Connection
	ch     *amqp.Channel
}

func (a *Amqp) Listen(writeChannel *chan messages.Message) {
	q := a.configureAmqp()
	a.consume(q, writeChannel)
}

func (a *Amqp) configureAmqp() (amqp.Queue) {
	var err error
	a.conn, err = amqp.Dial(a.Config.Url)
	failOnError(err, "Failed to connect to RabbitMQ")

	a.ch, err = a.conn.Channel()
	failOnError(err, "Failed to open channel")

	err = a.ch.ExchangeDeclare(
		a.Config.DeadLettersExchangeName,
		"fanout",
		true,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "Failed to declare dead letter exchange")

	dq, err := a.ch.QueueDeclare(
		a.Config.DeadLettersQueueName, // name
		false,                         // durable
		false,                         // delete when unused
		false,                         // exclusive
		false,                         // no-wait
		nil,
	)
	failOnError(err, "Failed to declare dead letter queue")

	err = a.ch.QueueBind(
		dq.Name,
		"#",
		a.Config.DeadLettersExchangeName,
		false,
		nil,
	)
	failOnError(err, "Failed to declare dead letter binding")

	err = a.ch.ExchangeDeclare(
		a.Config.ExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare exchange")

	q, err := a.ch.QueueDeclare(
		a.Config.QueueName, // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		amqp.Table{"x-dead-letter-exchange": "dead-letters"},
	)
	failOnError(err, "Failed to declare queue")

	err = a.ch.QueueBind(
		q.Name,
		"#",
		a.Config.ExchangeName,
		false,
		nil,
	)
	failOnError(err, "Failed to declare binding")

	err = a.ch.Qos(a.Config.QosPrefetchCount, 0, false)
	failOnError(err, "Failed setting Qos")

	return q
}

func (a *Amqp) consume(q amqp.Queue, writeChannel *chan messages.Message) {
	msgs, err := a.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	failOnError(err, "Failed to consume")

	defer a.conn.Close()
	defer a.ch.Close()

	go a.processMessages(msgs, writeChannel)
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {}
}

func (a *Amqp) processMessages(msgs <-chan amqp.Delivery, writeChannel *chan messages.Message) {
	msg := messages.Message{}
	for d := range msgs {
		msg.Id = d.DeliveryTag
		msg.Body = d.Body

		*writeChannel <- msg
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (a *Amqp) HandleAck(ackChannel *chan messages.Ack) {
	for ack := range *ackChannel {
		if ack.Ack {
			err := a.ch.Ack(ack.Id, false)
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

		err := a.ch.Nack(ack.Id, false, false)
		if err != nil {
			log.Fatal(err)
		}
	}
}
