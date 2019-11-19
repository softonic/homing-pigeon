package readers

import (
	"fmt"
	"strconv"
)

//import "github.com/streadway/amqp"


type AmqpReader struct {
	WriteChannel *chan string
	AckChannel *chan string
}

func (amqpReader *AmqpReader) Start() {
	go amqpReader.handleAck()
	for i := 0; i < 100; i++ {
		msg := "my message " + strconv.Itoa(i)
		fmt.Print("Sending message: " + msg + "\n")
		*amqpReader.WriteChannel <- msg
	}
}

func (amqpReader *AmqpReader) handleAck() {
	for ack := range *amqpReader.AckChannel {
		fmt.Print("Acked " + ack + "\n")
	}
}


//func (amqpReader *AmqpReader) Listen() {
//	conn, _ := amqp.Dial("amqp://ozjdarhq:s0xCJg7Njuok3T-Ghg3xZRzXKzUczq4C@lark.rmq.cloudamqp.com/ozjdarhq")
//	defer conn.Close()
//}
