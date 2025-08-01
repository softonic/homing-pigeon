package adapters

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/softonic/homing-pigeon/pkg/helpers"
	"github.com/softonic/homing-pigeon/pkg/messages"
	amqpAdapter "github.com/softonic/homing-pigeon/pkg/readers/adapters/amqp"
	"github.com/streadway/amqp"
	"k8s.io/klog"
)

type Amqp struct {
	ConsumedMessages <-chan amqp.Delivery
	Conn             amqpAdapter.Connection
	Ch               amqpAdapter.Channel
	Notify           chan *amqp.Error
}

func (a *Amqp) Listen(ctx context.Context, msgChannel chan<- messages.Message) {
	defer a.Conn.Close()
	defer a.Ch.Close()

	done := make(chan struct{})
	go func() {
		a.processMessages(ctx, msgChannel) // Pass context
		close(done)
	}()

	select {
	case <-ctx.Done(): // Graceful shutdown requested
		<-done // Wait for processMessages to finish
	case <-done: // processMessages finished naturally
	}
}

func (a *Amqp) processMessages(ctx context.Context, writeChannel chan<- messages.Message) {
	for {
		select {
		case <-ctx.Done():
			klog.V(4).Infoln("Context cancelled, stopping message processing.")
			return
		case err := <-a.Notify:
			if err != nil {
				klog.Fatalf("Error in connection: %s", err)
			}
			klog.V(4).Infoln("Closed connection.")
			break
		case d := <-a.ConsumedMessages:
			msg := messages.Message{}
			msg.Id = d.DeliveryTag
			msg.Body = d.Body

			writeChannel <- msg
		}
	}
}

func (a *Amqp) HandleAck(ackChannel <-chan messages.Message) {
	for ack := range ackChannel {
		if ack.IsAcked() {
			err := a.Ch.Ack(ack.Id, false)
			if err != nil {
				klog.Error(err)
			}
			continue
		}

		err := a.Ch.Nack(ack.Id, false, false)
		if err != nil {
			klog.Error(err)
		}
	}
}

// NewAMQPAdapter creates a new instance of AMQP adapter.
func NewAMQPAdapter() (ReadAdapter, error) {
	failOnError := func(err error, msg string) {
		if err != nil {
			klog.Errorf("%s: %s", msg, err)
		}
	}
	var err error
	var conn *amqp.Connection
	caPath := os.Getenv("RABBITMQ_CA_PATH")

	config, err := NewAmqpConfig()
	if err != nil {
		return nil, err
	}

	if caPath != "" {
		cfg := new(tls.Config)
		cfg.RootCAs = x509.NewCertPool()
		var ca []byte
		ca, err = os.ReadFile(caPath)
		if err == nil {
			cfg.RootCAs.AppendCertsFromPEM(ca)
			klog.V(0).Infof("Added CA certificate %s", caPath)
		}
		failOnError(err, "Failed loading RabbitMQ CA")

		tlsClientCert := os.Getenv("RABBITMQ_TLS_CLIENT_CERT")
		tlsClientKey := os.Getenv("RABBITMQ_TLS_CLIENT_KEY")
		if tlsClientCert != "" && tlsClientKey != "" {
			cert, err := tls.LoadX509KeyPair(tlsClientCert, tlsClientKey)
			if err == nil {
				cfg.Certificates = append(cfg.Certificates, cert)
				klog.V(0).Infof("Loaded RabbitMQ client cert %s", tlsClientCert)
				klog.V(0).Infof("Loaded RabbitMQ client key %s", tlsClientKey)
			}
			failOnError(err, "Failed loading RabbitMQ client certificate")
		}
		conn, err = amqp.DialTLS(config.Url, cfg)
		failOnError(err, "Failed to connect to RabbitMQ")
		if err == nil {
			klog.V(0).Infof("TLS Connection established")
		}
	} else {
		conn, err = amqp.Dial(config.Url)
		if err == nil {
			klog.V(0).Infof("Non TLS Connection established")
		}
	}
	failOnError(err, "Failed to connect to RabbitMQ")
	notify := conn.NotifyClose(make(chan *amqp.Error))

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
		config.DeadLettersQueueName,
		false,
		false,
		false,
		false,
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

	if config.OuterExchangeName != "" {
		err = ch.ExchangeDeclare(
			config.OuterExchangeName,
			config.OuterExchangeType,
			true,
			false,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare outer exchange")
		err = ch.ExchangeBind(
			config.ExchangeName,
			config.OuterExchangeBindingKey,
			config.OuterExchangeName,
			false,
			nil,
		)
		failOnError(err, "Failed to bind outer exchange")
	}

	queueArgs := make(amqp.Table)
	queueArgs["x-dead-letter-exchange"] = config.DeadLettersExchangeName
	if config.QueueMaxPriority > 0 {
		queueArgs["x-max-priority"] = config.QueueMaxPriority
	}
	q, err := ch.QueueDeclare(
		config.QueueName,
		false,
		false,
		false,
		false,
		queueArgs,
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
		q.Name,
		config.ConsumerName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		klog.Errorf("Failed to consume: %s", err)
	}

	return &Amqp{
		ConsumedMessages: msgs,
		Conn:             conn,
		Ch:               ch,
		Notify:           notify,
	}, nil
}

func generateConsumerID() (string, error) {
	consumerId := os.Getenv("CONSUMER_ID")
	if consumerId == "" {
		// Work out consumer ID based on hostname: useful for k8s resources (pods controlled by deployment, statefulset)
		hostname, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("could not get hostname for ConsumerID: %w", err)
		}
		pos := strings.LastIndex(hostname, "-")
		consumerId = hostname[pos+1:]
		if consumerId == "" {
			return "", fmt.Errorf("could not extract consumer ID from hostname: %s", hostname)
		}
	}

	return consumerId, nil
}
func NewAmqpConfig() (amqpAdapter.Config, error) {
	consumerId, err := generateConsumerID()
	if err != nil {
		return amqpAdapter.Config{}, fmt.Errorf("could not generate ConsumerID: %w", err)
	}
	klog.Infof("Generated consumer ID: %s", consumerId)

	queueName := strings.ReplaceAll(
		helpers.GetEnv("RABBITMQ_QUEUE_NAME", ""),
		"{{ .ConsumerId }}",
		consumerId,
	)

	qosPrefetchCount, err := strconv.Atoi(os.Getenv("RABBITMQ_QOS_PREFETCH_COUNT"))
	if err != nil {
		qosPrefetchCount = 0
	}

	QueueMaxPriority, err := strconv.Atoi(os.Getenv("RABBITMQ_QUEUE_MAX_PRIORITY"))
	if err != nil {
		QueueMaxPriority = 0
	}
	if QueueMaxPriority > 255 {
		klog.Warningf("RABBITMQ_QUEUE_MAX_PRIORITY is set to %d, but the maximum value is 255. Setting it to 255.", QueueMaxPriority)
		QueueMaxPriority = 255
	}

	consumerName := os.Getenv("RABBITMQ_CONSUMER_NAME")
	if consumerName == "" {
		consumerName, _ = os.Hostname()
	}

	return amqpAdapter.Config{
		Url:                     helpers.GetEnv("RABBITMQ_URL", ""),
		DeadLettersExchangeName: helpers.GetEnv("RABBITMQ_DLX_NAME", ""),
		DeadLettersQueueName:    helpers.GetEnv("RABBITMQ_DLX_QUEUE_NAME", ""),
		ExchangeName:            helpers.GetEnv("RABBITMQ_EXCHANGE_NAME", ""),
		ExchangeType:            helpers.GetEnv("RABBITMQ_EXCHANGE_TYPE", "fanout"),
		OuterExchangeName:       helpers.GetEnv("RABBITMQ_OUTER_EXCHANGE_NAME", ""),
		OuterExchangeType:       helpers.GetEnv("RABBITMQ_OUTER_EXCHANGE_TYPE", ""),
		OuterExchangeBindingKey: helpers.GetEnv("RABBITMQ_OUTER_EXCHANGE_BINDING_KEY", ""),
		QueueName:               queueName,
		QueueBindingKey:         helpers.GetEnv("RABBITMQ_QUEUE_BINDING_KEY", "#"),
		QosPrefetchCount:        qosPrefetchCount,
		QueueMaxPriority:        QueueMaxPriority,
		ConsumerName:            consumerName,
	}, nil
}
