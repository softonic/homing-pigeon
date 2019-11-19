package messages

type Message struct {
	Id uint64
	Body []byte
}

func (message Message) Nack() Ack{
	return Ack{
		Id: message.Id,
		Ack: false,
	}
}

func (message Message) Ack() Ack{
	return Ack{
		Id: message.Id,
		Ack: true,
	}
}
