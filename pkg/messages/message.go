package messages

import "errors"

type Message struct {
	Id    interface{}
	Body  []byte
	acked bool
}

func (m Message) Nack() (Message, error) {
	err := m.setAsAcked()
	if err != nil {
		return Message{}, err
	}

	return Message{
		Id:   m.Id,
		Body: []byte{0},
	}, nil
}

func (m Message) Ack() (Message, error) {
	err := m.setAsAcked()
	if err != nil {
		return Message{}, err
	}

	return Message{
		Id:   m.Id,
		Body: []byte{1},
	}, nil
}

func (m *Message) setAsAcked() error {
	if m.acked {
		return errors.New("Message already acked")
	}

	m.acked = true

	return nil
}
