package messages

type Ack struct {
	Id   interface{}
	Body []byte
	Ack  bool
}
