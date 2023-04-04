package main

import (
	"github.com/softonic/homing-pigeon/pkg/readers/adapters"
	"github.com/streadway/amqp"
	"k8s.io/klog"
	"log"
)

func main() {
	cfg, err := adapters.NewAmqpConfig()
	if err != nil {
		panic(err)
	}
	conn, err := amqp.Dial(cfg.Url)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		cfg.ExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare exchange")

	q, err := ch.QueueDeclare(
		cfg.QueueName, // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		amqp.Table{"x-dead-letter-exchange": "dead-letters"},
	)
	failOnError(err, "Failed to declare queue")

	err = ch.QueueBind(
		q.Name,
		"#",
		cfg.ExchangeName,
		false,
		nil,
	)
	failOnError(err, "Failed to declare binding")
	for {
		err = ch.Publish(cfg.ExchangeName, "#", false, false, amqp.Publishing{Body: []byte("{\"meta\": {\"index\":{\"_index\":\"test\"}},\"data\": {\"field1\":\"value1\"}}")})
		if err != nil {
			log.Printf("Message not published: %v", err)
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		klog.Errorf("%s: %s", msg, err)
	}
}
