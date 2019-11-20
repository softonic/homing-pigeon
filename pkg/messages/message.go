package messages

import "errors"

type Message struct {
	Id uint64
	Body []byte
	acked bool
}

func (m Message) Nack() (*Ack, error) {
	err := m.setAsAcked()
	if err != nil {
		return nil, err
	}

	return &Ack{
		Id:  m.Id,
		Ack: false,
	}, nil
}

func (m Message) Ack() (*Ack, error){
	err := m.setAsAcked()
	if err != nil {
		return nil, err
	}

	return &Ack{
		Id:  m.Id,
		Ack: true,
	}, nil
}

func (m Message) setAsAcked() error {
	if m.acked {
		return errors.New("Message already acked")
	}

	m.acked = true

	return nil
}
