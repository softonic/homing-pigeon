package writers

import "fmt"

type ElasticsearchWriter struct {
	WriteChannel *chan string
	AckChannel *chan string
	acks []string
	internalAck *chan string
}

func (elasticsearchWriter *ElasticsearchWriter) Start() {
	c := make(chan string)
	elasticsearchWriter.internalAck = &c
	elasticsearchWriter.acks = make([]string, 0)
	go elasticsearchWriter.timeout()
	count := 0
	for msg := range *elasticsearchWriter.WriteChannel {
		fmt.Print(msg)
		count++
		*elasticsearchWriter.internalAck <- "ack " + msg
		if count > 50 {
			elasticsearchWriter.handleAcks()
		}
	}
	// Insert bulk
	// on success:

}

func (elasticsearchWriter *ElasticsearchWriter) handleAcks() {
	for range *elasticsearchWriter.internalAck {
		for _, ack := range elasticsearchWriter.acks {
			*elasticsearchWriter.AckChannel <- ack
		}
		elasticsearchWriter.acks = make([]string, 0)
	}
}

func (elasticsearchWriter *ElasticsearchWriter) timeout() {
	// for { sleep(N); handleAcks }
	// *elasticsearchWriter.internalAck <- "now"
}
