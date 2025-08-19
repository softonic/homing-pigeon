package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"github.com/softonic/homing-pigeon/pkg/messages"
	esAdapter "github.com/softonic/homing-pigeon/pkg/writers/adapters/elasticsearch"
	"k8s.io/klog"
)

type Elasticsearch struct {
	FlushMaxSize  int
	FlushInterval time.Duration
	Bulk          esapi.Bulk
}

func (es *Elasticsearch) ProcessMessages(msgs *[]messages.Message) {

	if len(*msgs) == 0 {
		return
	}

	var buf bytes.Buffer

	for i := range *msgs {
		msg := &(*msgs)[i]
		// Process all messages regardless of ack/nack status
		body, err := es.decodeBody(msg.Body)
		if err != nil {
			klog.Errorf("Invalid Message: %s", string(msg.Body))
			msg.Nack()
			continue
		}
		err = es.writeToBuffer(&buf, body)
		if err != nil {
			continue
		}
	}

	if buf.Len() == 0 {
		return
	}
	result, err := es.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil || result.IsError() {
		klog.Warningf("Error in bulk action, %v", err)
		es.setAllNacks(msgs)
		return
	}

	klog.V(4).Infof("Elasticsearch bulk request successful, processing %d items", len(*msgs))
	response := es.getResponseFromResult(result)
	result.Body.Close()
	es.setAcksFromResponse(response, msgs)
}

func (es *Elasticsearch) setAcksFromResponse(response esAdapter.ElasticSearchBulkResponse, msgs *[]messages.Message) {
	maxValidStatus := 299

	responseItemPos := 0
	for i := range *msgs {
		msg := &(*msgs)[i]

		item := response.Items[responseItemPos].(map[string]interface{})
		for _, data := range item {
			values := data.(map[string]interface{})
			status := int(values["status"].(float64))

			if status > maxValidStatus {
				klog.Warningf("Item has invalid status: %v", data)

				msg.Nack()
			} else {
				msg.Ack()
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
		klog.Errorf("Error in elasticsearch response: %v %v", err, response)
	}
	return response
}

func (es *Elasticsearch) setAllNacks(msgs *[]messages.Message) {
	for i := range *msgs {
		(*msgs)[i].Nack()
	}
}

func (es *Elasticsearch) writeToBuffer(buf *bytes.Buffer, body esAdapter.ElasticsearchBody) error {
	meta, err := json.Marshal(body.Meta)
	if err != nil {
		return err
	}
	if bytes.Equal(meta, []byte("null")) {
		return errors.New("invalid body: meta should be present")
	}
	data, err := json.Marshal(body.Data)
	if err != nil {
		return err
	}

	payload := append(meta, "\n"...)
	if !bytes.Equal(data, []byte("null")) {
		payload = append(payload, data...)
		payload = append(payload, "\n"...)
	}

	buf.Write(payload)

	return nil
}

func (es *Elasticsearch) ShouldProcess(msgs []messages.Message) bool {
	return len(msgs) >= es.FlushMaxSize
}

func (es *Elasticsearch) GetTimeout() time.Duration {
	return es.FlushInterval
}

func (es *Elasticsearch) decodeBody(msg []byte) (esAdapter.ElasticsearchBody, error) {
	body := esAdapter.ElasticsearchBody{}
	err := json.Unmarshal(msg, &body)

	return body, err
}

func NewElasticsearchAdapter() (WriteAdapter, error) {
	flushMaxSize, err := strconv.Atoi(os.Getenv("ELASTICSEARCH_FLUSH_MAX_SIZE"))
	if err != nil {
		flushMaxSize = 1
	}

	flushMaxIntervalMs, err := strconv.Atoi(os.Getenv("ELASTICSEARCH_FLUSH_MAX_INTERVAL_MS"))
	if err != nil {
		flushMaxIntervalMs = 1000
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{})
	if err != nil {
		return nil, err
	}
	return &Elasticsearch{
		FlushMaxSize:  flushMaxSize,
		FlushInterval: time.Duration(flushMaxIntervalMs) * time.Millisecond,
		Bulk:          es.Bulk,
	}, nil
}
