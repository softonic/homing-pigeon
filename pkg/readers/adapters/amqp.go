package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/streadway/amqp"
	"log"
)

type Amqp struct {
	ch *amqp.Channel
}

func (a *Amqp) Listen(writeChannel *chan messages.Message) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit-mq:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	a.ch, err = conn.Channel()
	failOnError(err, "Failed to open channel")
	defer a.ch.Close()

	err = a.ch.ExchangeDeclare(
		"dead-letters",
		"fanout",
		true,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "Failed to declare dead letter exchange")

	dq, err := a.ch.QueueDeclare(
		"dead", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,
	)
	failOnError(err, "Failed to declare dead letter queue")

	err = a.ch.QueueBind(
		dq.Name,
		"#",
		"dead-letters",
		false,
		nil,
	)
	failOnError(err, "Failed to declare dead letter binding")

	err = a.ch.ExchangeDeclare(
		"hello",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare exchange")

	q, err := a.ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		amqp.Table{"x-dead-letter-exchange": "dead-letters"},
	)
	failOnError(err, "Failed to declare queue")

	err = a.ch.QueueBind(
		q.Name,
		"#",
		"hello",
		false,
		nil,
	)
	failOnError(err, "Failed to declare binding")

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

	err = a.ch.Qos(500, 0, false)
	failOnError(err, "Failed setting Qos")

	forever := make(chan bool)

	go func() {
		msg := messages.Message{}
		for d := range msgs {
			msg.Id = d.DeliveryTag
			msg.Body = d.Body

			*writeChannel <- msg
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
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
		} else {
			err := a.ch.Nack(ack.Id, false, false)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
