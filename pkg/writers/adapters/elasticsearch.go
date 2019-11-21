package adapters

import (
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/softonic/homing-pigeon/pkg/messages"
	esAdapter "github.com/softonic/homing-pigeon/pkg/writers/adapters/elasticsearch"
	"log"
)

type Elasticsearch struct{
	FlushMaxSize       int
	FlushMaxIntervalMs int64
}

func (es *Elasticsearch) ProcessMessages(msgs []*messages.Message) []*messages.Ack {
	acks := make([]*messages.Ack, len(msgs))

	if len(msgs) == 0 {
		return acks
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer

	for i, msg := range msgs {
		body, err := es.decodeBody(msg.Body)
		if err != nil {
			log.Printf("Invalid Message: %v", *msg)
			nack, err := msg.Nack()
			if err != nil {
				log.Fatal(err)
			}
			acks[i] = nack
			continue
		}
		err = es.writeToBuffer(&buf, body)
		if err != nil {
			continue
		}
	}

	if buf.Len() == 0 {
		return acks
	}

	result, err := client.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil || result.IsError() {
		log.Printf("Error in bulk action, %v", err)
		es.setAllNacks(msgs, acks)
		return acks
	}

	response := es.getResponseFromResult(result)
	es.setAcksFromResponse(response, msgs, acks)

	return acks
}

func (es *Elasticsearch) setAcksFromResponse(response esAdapter.ElasticSearchBulkResponse, msgs []*messages.Message, acks []*messages.Ack) {
	log.Printf("Result: %v", response)
	maxValidStatus := 299

	responseItemPos := 0
	for ackPos, ack := range acks {
		if ack != nil {
			continue
		}

		item := response.Items[responseItemPos].(map[string]interface{})
		for _, data := range item {
			values := data.(map[string]interface{})
			status := int(values["status"].(float64))

			if status > maxValidStatus {
				log.Printf("NACK: %v", *msgs[ackPos])
				ack, err := msgs[ackPos].Nack()
				if err == nil {
					acks[ackPos] = ack
				}
			} else {
				log.Printf("ACK: %v", *msgs[ackPos])
				ack, err := msgs[ackPos].Ack()
				if err == nil {
					acks[ackPos] = ack
				}
			}
		}
		responseItemPos++

	}
}

func (es *Elasticsearch) getResponseFromResult(result *esapi.Response) esAdapter.ElasticSearchBulkResponse {
	response := esAdapter.ElasticSearchBulkResponse{}
	d := json.NewDecoder(result.Body)
	err := d.Decode(&response)
	if err != nil {
		log.Fatalf("Error in elasticsearch response: %v %v", err, response)
	}
	return response
}

func (es *Elasticsearch) setAllNacks(msgs []*messages.Message, acks []*messages.Ack) {
	for i, msg := range msgs {
		nack, err := msg.Nack()
		if err == nil {
			acks[i] = nack
		}
	}
}

func (es *Elasticsearch) writeToBuffer(buf *bytes.Buffer, body esAdapter.ElasticsearchBody) error {
	meta, err := json.Marshal(body.Meta)
	if err != nil {
		return err
	}
	data, err := json.Marshal(body.Data)
	if err != nil {
		return err
	}

	payload := append(meta, "\n"...)
	payload = append(payload, data...)
	payload = append(payload, "\n"...)

	buf.Write(payload)

	return nil
}

func (es *Elasticsearch) ShouldProcess(msgs []*messages.Message) bool {
	return len(msgs) >= es.FlushMaxSize
}

func (es *Elasticsearch) GetTimeoutInMs() int64 {
	return es.FlushMaxIntervalMs
}

func (es *Elasticsearch) decodeBody(msg []byte) (esAdapter.ElasticsearchBody, error) {
	body := esAdapter.ElasticsearchBody{}
	err := json.Unmarshal(msg, &body)

	return body, err
}
