package amqp

type Connection interface {
	Close() error
}