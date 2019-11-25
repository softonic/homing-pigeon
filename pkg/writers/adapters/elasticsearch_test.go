package adapters

import (
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func mockBulkFunc() esapi.Bulk {
	return func(body io.Reader, o ...func(*esapi.BulkRequest)) (*esapi.Response, error) {
		return nil, nil
	}
}


func TestAdapterProcessSingleMessage(t *testing.T) {
	if err != nil {
		assert.Fail(t, "Elasticsearch client could not be mock", err)
	}

	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          mockBulkFunc(),
	}

	
}