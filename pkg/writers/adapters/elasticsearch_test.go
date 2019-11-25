package adapters

import (
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"strings"
	"testing"
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
		args := b.Called(body)

		var err error
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

	acks := esAdapter.ProcessMessages([]*messages.Message{
		{
			Id:   0,
			Body: []byte("{ Invalid Json }"),
		},
	})

	bulk.AssertNotCalled(t, "func1")
	assert.Len(t, acks, 1)
	assert.False(t, acks[0].Ack)
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

	acks := esAdapter.ProcessMessages([]*messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"valid\": \"json\" }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.False(t, acks[0].Ack)
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
		Body:       nil,
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	bulk.SetBody("{\"errors\":false,\"items\":[{\"create\":{\"status\":200}}]}")
	func (b *BulkMock) SetBody(s string) {
		b.Body = ioutil.NopCloser(strings.NewReader(s))
	}

	acks := esAdapter.ProcessMessages([]*messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"valid\": \"json\" }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.True(t, acks[0].Ack)
}
