package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io"
)

type BulkClient interface {
	Bulk(body io.Reader, o ...func(request *esapi.BulkRequest)) (*esapi.Response, error)
}
