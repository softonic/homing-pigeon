package amqp

type Config struct {
	Url                     string
	DeadLettersExchangeName string
	DeadLettersQueueName    string
	ExchangeName            string
	QueueName               string
	QosPrefetchCount        int
	ConsumerName            string
	ExchangeType            string
	QueueBindingKey         string
}