package adapters

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/softonic/homing-pigeon/mocks"
	"testing"
)

type mockEsClient struct {
	elasticsearch.Client
}

func TestAdapterProcessSingleMessage(t *testing.T) {
	mockBulkClient := new(mocks.BulkClient)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Client:        mockBulkClient,
	}
}
