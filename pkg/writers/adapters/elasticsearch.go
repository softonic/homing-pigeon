package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"log"
)

type ElasticsearchBody struct {
	Meta interface{} `json:"meta"`
	Data interface{} `json:"data"`
}
type Elasticsearch struct{}

func (wa *Elasticsearch) ProcessMessages(msgs []messages.Message) []messages.Ack {
	acks := make([]messages.Ack, 0)

	if len(msgs) == 0 {
		return acks
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elasticsearch:9200",
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer

	for _, msg := range msgs {
		body, err := wa.decodeBody(msg.Body)
		if err != nil {
			log.Fatal("Invalid Message")
			acks = append(acks, msg.Nack())
			continue
		}

		acks, err = wa.writeToBuffer(&buf, body.Meta, &msg, acks)
		if err != nil {
			continue
		}

		acks, err = wa.writeToBuffer(&buf, body.Data, &msg, acks)
		if err != nil {
			continue
		}

		// Temporary
		acks = append(acks, msg.Ack())
	}

	log.Print(buf.Bytes())
	result, err := client.Bulk(bytes.NewReader(buf.Bytes()))
	// check result, check failed records and add to ack
	if err != nil {
		fmt.Println(err)
	}

	log.Print(result)
	log.Printf("%s", buf.Bytes())

	return acks
}

func (wa *Elasticsearch) writeToBuffer(buf *bytes.Buffer, data interface{}, msg *messages.Message, acks []messages.Ack) ([]messages.Ack, error) {
	json, err := json.Marshal(data)
	if err != nil {
		return append(acks, msg.Nack()), err
	}

	buf.Write(append(json, "\n"...))

	return acks, nil
}

func (wa *Elasticsearch) ShouldProcess(msgs []messages.Message) bool {
	return len(msgs) > 20
}

func (wa *Elasticsearch) GetTimeoutInMs() int64 {
	return int64(5000)
}

func (wa *Elasticsearch) decodeBody(msg []byte) (ElasticsearchBody, error) {
	body := ElasticsearchBody{}
	err := json.Unmarshal(msg, &body)

	log.Printf("Message: %s", msg)
	log.Printf("Body: %+v", body)

	return body, err
}
