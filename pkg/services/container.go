package services

import (
	"bytes"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sarulabs/dingo"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/pkg/readers"
	readAdapters "github.com/softonic/homing-pigeon/pkg/readers/adapters"
	amqpAdapter "github.com/softonic/homing-pigeon/pkg/readers/adapters/amqp"
	"github.com/softonic/homing-pigeon/pkg/writers"
	writeAdapters "github.com/softonic/homing-pigeon/pkg/writers/adapters"
	"github.com/streadway/amqp"
	"html/template"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var Container = []dingo.Def{
	{
		Name: "Reader",
		Build: func(msgChannel chan messages.Message, ackChannel chan messages.Ack, readAdapter readAdapters.ReadAdapter) (*readers.Reader, error) {
			return &readers.Reader{
				ReadAdapter: readAdapter,
				MsgChannel:  msgChannel,
				AckChannel:  ackChannel,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("MsgChannel"),
			"1": dingo.Service("AckChannel"),
			"2": dingo.Service("AmqpAdapter"),
		},
	},
	{
		Name: "DummyAdapter",
		Build: func() (readAdapters.ReadAdapter, error) {
			return &readAdapters.Dummy{}, nil
		},
	},
	{
		Name: "AmqpAdapter",
		Build: func(config amqpAdapter.Config, ) (readAdapters.ReadAdapter, error) {
			failOnError := func(err error, msg string) {
				if err != nil {
					log.Fatalf("%s: %s", msg, err)
				}
			}

			conn, err := amqp.Dial(config.Url)
			failOnError(err, "Failed to connect to RabbitMQ")

			ch, err := conn.Channel()
			failOnError(err, "Failed to open channel")

			err = ch.ExchangeDeclare(
				config.DeadLettersExchangeName,
				"fanout",
				true,
				false,
				true,
				false,
				nil,
			)
			failOnError(err, "Failed to declare dead letter exchange")

			dq, err := ch.QueueDeclare(
				config.DeadLettersQueueName, // name
				false,                       // durable
				false,                       // delete when unused
				false,                       // exclusive
				false,                       // no-wait
				nil,
			)
			failOnError(err, "Failed to declare dead letter queue")

			err = ch.QueueBind(
				dq.Name,
				"#",
				config.DeadLettersExchangeName,
				false,
				nil,
			)
			failOnError(err, "Failed to declare dead letter binding")

			internalExchange := os.Getenv("RABBITMQ_EXCHANGE_INTERNAL")

			isInternalExchange := false
			if internalExchange == "true" {
				isInternalExchange = true
			}

			err = ch.ExchangeDeclare(
				config.ExchangeName,
				config.ExchangeType,
				true,
				false,
				isInternalExchange,
				false,
				nil,
			)
			failOnError(err, "Failed to declare exchange")

			if config.ExchangeBindingName != "" {
				ch.ExchangeBind(
					config.ExchangeBindingName,
					config.ExchangeBindingKey,
					config.ExchangeName,
					false,
					nil,
				)
			}
			q, err := ch.QueueDeclare(
				config.QueueName, // name
				false,            // durable
				false,            // delete when unused
				false,            // exclusive
				false,            // no-wait
				amqp.Table{"x-dead-letter-exchange": config.DeadLettersExchangeName},
			)
			failOnError(err, "Failed to declare queue")

			err = ch.QueueBind(
				q.Name,
				config.QueueBindingKey,
				config.ExchangeName,
				false,
				nil,
			)
			failOnError(err, "Failed to declare binding")

			err = ch.Qos(config.QosPrefetchCount, 0, false)
			failOnError(err, "Failed setting Qos")

			msgs, err := ch.Consume(
				q.Name,              // queue
				config.ConsumerName, // consumer
				false,               // auto-ack
				false,               // exclusive
				false,               // no-local
				false,               // no-wait
				nil,
			)
			if err != nil {
				log.Fatalf("Failed to consume: %s", err)
			}

			return &readAdapters.Amqp{
				ConsumedMessages: msgs,
				Conn:             conn,
				Ch:               ch,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("AmqpConfig"),
		},
	},
	{
		Name: "AmqpConfig",
		Build: func() (amqpAdapter.Config, error) {
			// @TODO this needs to be extracted to its own method
			getEnv := func(envVarName string, defaultValue string) string {
				value := os.Getenv(envVarName)
				if value != "" {
				return value
			}
				return defaultValue
			}

			// @TODO this needs to be extracted to its own method
			consumerId := os.Getenv("CONSUMER_ID")
			if consumerId == "" {
				// Work out consumer ID based on hostname: useful for k8s resources (pods controlled by deployment, statefulset)
				hostname, err := os.Hostname()
				if err != nil {
					log.Fatalf("Could not set ConsumerID: %v", err)
				}
				pos := strings.LastIndex(hostname, "-")
				consumerId = hostname[pos+1:]
			}

			// @TODO This needs to be extracted to its own object
			data := struct {
				ConsumerId string
			}{
				consumerId,
			}

			tpl := template.New("queueName")
			tpl, err := tpl.Parse(getEnv("RABBITMQ_QUEUE_NAME", ""))
			if err != nil {
				log.Fatalf("Invalid RABBITMQ_QUEUE_NAME: %v", err)
			}
			var buf bytes.Buffer
			err = tpl.Execute(&buf, data)
			if err != nil {
				log.Fatalf("Invalid RABBITMQ_QUEUE_NAME: %v", err)
			}
			queueName := buf.String()

			qosPrefetchCount, err := strconv.Atoi(os.Getenv("RABBITMQ_QOS_PREFETCH_COUNT"))
			if err != nil {
				qosPrefetchCount = 0
			}

			consumerName := os.Getenv("RABBITMQ_CONSUMER_NAME")
			if consumerName == "" {
				consumerName, _ = os.Hostname()
			}

			return amqpAdapter.Config{
				Url:                     getEnv("RABBITMQ_URL", ""),
				DeadLettersExchangeName: getEnv("RABBITMQ_DLX_NAME", ""),
				DeadLettersQueueName:    getEnv("RABBITMQ_DLX_QUEUE_NAME", ""),
				ExchangeName:            getEnv("RABBITMQ_EXCHANGE_NAME", ""),
				ExchangeType:            getEnv("RABBITMQ_EXCHANGE_TYPE", "fanout"),
				ExchangeBindingName:     getEnv("RABBITMQ_EXCHANGE_BINDING_NAME", ""),
				ExchangeBindingKey:      getEnv("RABBITMQ_EXCHANGE_BINDING_KEY", ""),
				QueueName:               queueName,
				QueueBindingKey:         getEnv("RABBITMQ_QUEUE_BINDING_KEY", "#"),
				QosPrefetchCount:        qosPrefetchCount,
				ConsumerName:            consumerName,
			}, nil
		},
	},
	{
		Name: "Writer",
		Build: func(msgChannel chan messages.Message, ackChannel chan messages.Ack, writeAdapter writeAdapters.WriteAdapter) (*writers.Writer, error) {

			return &writers.Writer{
				MsgChannel:   msgChannel,
				AckChannel:   ackChannel,
				WriteAdapter: writeAdapter,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("MsgChannel"),
			"1": dingo.Service("AckChannel"),
			"2": dingo.Service("ElasticsearchAdapter"),
		},
	},

	{
		Name: "ElasticsearchAdapter",
		Build: func(bulk esapi.Bulk) (writeAdapters.WriteAdapter, error) {
			flushMaxSize, err := strconv.Atoi(os.Getenv("ELASTICSEARCH_FLUSH_MAX_SIZE"))
			if err != nil {
				flushMaxSize = 1
			}

			flushMaxIntervalMs, err := strconv.Atoi(os.Getenv("ELASTICSEARCH_FLUSH_MAX_INTERVAL_MS"))
			if err != nil {
				flushMaxIntervalMs = 1000
			}
			return &writeAdapters.Elasticsearch{
				FlushMaxSize:  flushMaxSize,
				FlushInterval: time.Duration(flushMaxIntervalMs) * time.Millisecond,
				Bulk:          bulk,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("ElasticsearchBulkClient"),
		},
	},
	{
		Name: "ElasticsearchBulkClient",
		Build: func() (esapi.Bulk, error) {
			es, err := elasticsearch.NewClient(elasticsearch.Config{})
			if err != nil {
				return nil, err
			}

			return es.Bulk, nil
		},
	},
	{
		Name: "NopAdapter",
		Build: func() (writeAdapters.WriteAdapter, error) {
			return &writeAdapters.Nop{}, nil
		},
	},
	{
		Name: "MsgChannel",
		Build: func() (chan messages.Message, error) {
			bufLen, err := strconv.Atoi(os.Getenv("MESSAGE_BUFFER_LENGTH"))
			if err != nil {
				bufLen = 0
			}
			c := make(chan messages.Message, bufLen)
			return c, nil
		},
	},
	{
		Name: "AckChannel",
		Build: func() (chan messages.Ack, error) {
			bufLen, err := strconv.Atoi(os.Getenv("ACK_BUFFER_LENGTH"))
			if err != nil {
				bufLen = 0
			}
			c := make(chan messages.Ack, bufLen)
			return c, nil
		},
	},
}
