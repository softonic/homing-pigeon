package amqp

type Channel interface {
	Close() error
	Ack(tag uint64, multiple bool) error
	Nack(tag uint64, multiple bool, requeue bool) error
}

