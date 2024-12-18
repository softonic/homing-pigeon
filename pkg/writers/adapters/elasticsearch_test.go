package adapters

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/**
 * Manual mock
 *
 * This is done in a manual way because the elasticsearch client does not implement
 * any interface, so we cannot mock it directly.
 */
type BulkMock struct {
	mock.Mock
}

func (b *BulkMock) getBulkFunc() esapi.Bulk {
	return func(body io.Reader, o ...func(*esapi.BulkRequest)) (*esapi.Response, error) {
		var err error
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(body)
		if err != nil {
			panic(err)
		}
		args := b.Called(buf.String())
		err = nil
		if args.Get(1) != nil {
			err = args.Get(0).(error)
		}

		return args.Get(0).(*esapi.Response), err
	}
}

func TestAdapterReceiveInvalidMessage(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	msgs := []messages.Message{
		{
			Id:   1,
			Body: []byte("{ Invalid Json }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertNotCalled(t, "func1")
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsNacked())
}

func TestBulkActionWithErrorsMustDiscardAllMessages(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 404,
		Header:     nil,
		Body:       nil,
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsNacked())
}

func TestBulkActionWithSingleItemSucessful(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":false,\"items\":[{\"create\":{\"status\":200}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsAcked())
}

func TestBulkActionWithSingleItemUnsuccessful(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"create\":{\"status\":409}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsNacked())
}

func TestBulkActionWithMixedItemStatus(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"create\":{\"status\":409}},{\"create\":{\"status\":200}},{\"create\":{\"status\":409}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
		{
			Id:   1,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
		{
			Id:   2,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 3)
	assert.True(t, msgs[0].IsNacked())
	assert.True(t, msgs[1].IsAcked())
	assert.True(t, msgs[2].IsNacked())
}

func TestBulkActionWithOnlyMetadata(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":false,\"items\":[{\"delete\":{\"status\":200}}]}")),
	}
	expectedBody := "{\"delete\":{\"_id\":\"123\"}}\n"
	bulk.On("func1", expectedBody).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"123\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsAcked())
}

func TestBulkActionWithNoMetadata(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"foobar\": {\"delete\": {\"_id\":\"123\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertNotCalled(t, "func1", mock.Anything)
	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.Empty(t, msgs[0].IsNacked())
}
