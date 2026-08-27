package adapters

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/esapi"
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
	assert.True(t, msgs[0].IsNacked())
}

func TestMessageWithoutMetadataInBatchDoesNotShiftResponseAttribution(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	// Only the first and last messages reach the bulk request, so the
	// response carries two items: a failure for msg 0 and a success for
	// msg 2. The meta-less msg 1 must not consume a response position.
	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"create\":{\"_id\":\"1\",\"status\":409}},{\"create\":{\"_id\":\"3\",\"status\":200}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"create\": {\"_id\":\"1\"}} }"),
		},
		{
			Id:   1,
			Body: []byte("{ \"foobar\": {\"create\": {\"_id\":\"2\"}} }"),
		},
		{
			Id:   2,
			Body: []byte("{ \"meta\": {\"create\": {\"_id\":\"3\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 3)
	assert.True(t, msgs[0].IsNacked())
	assert.True(t, msgs[1].IsNacked())
	assert.True(t, msgs[2].IsAcked())
}

func TestDeleteNotFoundIsNackedByDefault(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":false,\"items\":[{\"delete\":{\"_id\":\"123\",\"result\":\"not_found\",\"status\":404}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"123\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsNacked())
}

func TestDeleteNotFoundIsAckedWhenAckDeleteNotFoundIsEnabled(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:      0,
		FlushInterval:     0,
		Bulk:              bulk.getBulkFunc(),
		AckDeleteNotFound: true,
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":false,\"items\":[{\"delete\":{\"_id\":\"123\",\"result\":\"not_found\",\"status\":404}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

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

func TestDeleteNotFoundWithErrorIsNackedWhenAckDeleteNotFoundIsEnabled(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:      0,
		FlushInterval:     0,
		Bulk:              bulk.getBulkFunc(),
		AckDeleteNotFound: true,
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"delete\":{\"_id\":\"123\",\"error\":{\"type\":\"index_not_found_exception\",\"reason\":\"no such index [foo]\"},\"status\":404}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"123\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsNacked())
}

func TestNonDeleteNotFoundIsNackedWhenAckDeleteNotFoundIsEnabled(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:      0,
		FlushInterval:     0,
		Bulk:              bulk.getBulkFunc(),
		AckDeleteNotFound: true,
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"index\":{\"_id\":\"123\",\"status\":404}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"index\": {\"_id\":\"123\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 1)
	assert.True(t, msgs[0].IsNacked())
}

func TestBulkActionWithMixedItemStatusAndAckDeleteNotFoundEnabled(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:      0,
		FlushInterval:     0,
		Bulk:              bulk.getBulkFunc(),
		AckDeleteNotFound: true,
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"delete\":{\"_id\":\"1\",\"result\":\"not_found\",\"status\":404}},{\"create\":{\"_id\":\"2\",\"status\":409}},{\"create\":{\"_id\":\"3\",\"status\":200}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"1\"}} }"),
		},
		{
			Id:   1,
			Body: []byte("{ \"meta\": {\"create\": {\"_id\":\"2\"}} }"),
		},
		{
			Id:   2,
			Body: []byte("{ \"meta\": {\"create\": {\"_id\":\"3\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 3)
	assert.True(t, msgs[0].IsAcked())
	assert.True(t, msgs[1].IsNacked())
	assert.True(t, msgs[2].IsAcked())
}

func TestUndecodableBulkResponseNacksAllMessagesWithoutPanicking(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 200,
		Header:     nil,
		Body:       io.NopCloser(strings.NewReader("this-is-not-json")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	msgs := []messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"1\"}} }"),
		},
		{
			Id:   1,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"2\"}} }"),
		},
	}

	esAdapter.ProcessMessages(&msgs)

	bulk.AssertExpectations(t)
	assert.Len(t, msgs, 2)
	assert.True(t, msgs[0].IsNacked())
	assert.True(t, msgs[1].IsNacked())
}

func TestNewElasticsearchAdapterReadsAckDeleteNotFoundFromEnv(t *testing.T) {
	t.Setenv("ELASTICSEARCH_ACK_DELETE_NOT_FOUND", "true")

	adapter, err := NewElasticsearchAdapter()

	assert.NoError(t, err)
	assert.True(t, adapter.(*Elasticsearch).AckDeleteNotFound)
}

func TestNewElasticsearchAdapterDefaultsAckDeleteNotFoundToFalse(t *testing.T) {
	adapter, err := NewElasticsearchAdapter()

	assert.NoError(t, err)
	assert.False(t, adapter.(*Elasticsearch).AckDeleteNotFound)
}

func TestShouldProcessWhenTheByteBudgetIsReachedBeforeTheMessageCount(t *testing.T) {
	// Elasticsearch rejects a bulk request over http.max_content_length, and
	// FlushMaxSize alone cannot prevent it: 800 messages of a few hundred KB
	// each are far past any sane limit. The byte budget has to close the batch
	// first.
	es := &Elasticsearch{FlushMaxSize: 800, FlushMaxBytes: 250}

	msgs := []messages.Message{
		{Id: 1, Body: make([]byte, 100)},
		{Id: 2, Body: make([]byte, 100)},
		{Id: 3, Body: make([]byte, 100)},
	}

	assert.False(t, es.ShouldProcess(msgs[:2]), "200 bytes is still under the budget")
	assert.True(t, es.ShouldProcess(msgs), "300 bytes is over the budget")
}

func TestShouldProcessIgnoresTheByteBudgetWhenItIsNotSet(t *testing.T) {
	es := &Elasticsearch{FlushMaxSize: 2, FlushMaxBytes: 0}

	msgs := []messages.Message{
		{Id: 1, Body: make([]byte, 10_000)},
	}

	assert.False(t, es.ShouldProcess(msgs), "with no byte budget only the message count counts")
}

func TestNewElasticsearchAdapterReadsFlushMaxBytesFromEnv(t *testing.T) {
	t.Setenv("ELASTICSEARCH_FLUSH_MAX_BYTES", "1234")

	adapter, err := NewElasticsearchAdapter()

	assert.NoError(t, err)
	assert.Equal(t, 1234, adapter.(*Elasticsearch).FlushMaxBytes)
}

func TestNewElasticsearchAdapterDefaultsFlushMaxBytesToEightyMegabytes(t *testing.T) {
	adapter, err := NewElasticsearchAdapter()

	assert.NoError(t, err)
	assert.Equal(t, 80*1024*1024, adapter.(*Elasticsearch).FlushMaxBytes)
}
