package main

import (
	"github.com/streadway/amqp"
	"log"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbit-mq:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"hello",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare exchange")

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		amqp.Table{"x-dead-letter-exchange": "dead-letters"},
	)
	failOnError(err, "Failed to declare queue")

	err = ch.QueueBind(
		q.Name,
		"#",
		"hello",
		false,
		nil,
	)
	failOnError(err, "Failed to declare binding")
	for {
		err = ch.Publish("hello", "#", false, false, amqp.Publishing{Body: []byte("{\"meta\": {\"index\":{\"_index\":\"test\"}},\"data\": {\"field1\":\"value1\"}}")})
		if err != nil {
			log.Printf("Message not published: %v", err)
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}