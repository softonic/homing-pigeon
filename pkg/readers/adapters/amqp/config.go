package amqp

type Config struct {
	Url                     string
	DeadLettersExchangeName string
	DeadLettersQueueName    string
	ExchangeName            string
	QueueName               string
	QueueMaxPriority        int
	QosPrefetchCount        int
	ConsumerName            string
	ExchangeType            string
	QueueBindingKey         string
	OuterExchangeName       string
	OuterExchangeBindingKey string
	OuterExchangeType       string
}
