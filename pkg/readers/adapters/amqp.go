package adapters

import (
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/streadway/amqp"
	"log"
)

type Amqp struct{
	ch *amqp.Channel
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (a *Amqp) Listen(writeChannel *chan messages.Message) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit-mq:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	a.ch, err = conn.Channel()
	failOnError(err, "Failed to open channel")
	defer a.ch.Close()

	q, err := a.ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare queue")

	msgs, err := a.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to consume")

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

func (a *Amqp) HandleAck(ackChannel *chan messages.Ack) {
	for ack := range *ackChannel {
		err := a.ch.Ack(ack.Id, false)
		if err != nil {
			log.Fatal(err)
		}
	}
}
