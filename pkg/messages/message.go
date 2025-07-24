package messages

type Message struct {
	Id    uint64
	Body  []byte
	acked *bool
}

func (m *Message) IsAcked() bool {
	return m.acked != nil && *m.acked
}

func (m *Message) IsNacked() bool {
	return m.acked != nil && !*m.acked
}

func (m *Message) Nack() {
	m.acked = func(b bool) *bool { return &b }(false)
}

func (m *Message) Ack() {
	m.acked = func(b bool) *bool { return &b }(true)
}
