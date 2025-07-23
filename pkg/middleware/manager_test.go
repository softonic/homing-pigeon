package middleware

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// MockMiddlewareServer implements a dummy gRPC middleware service for testing
type MockMiddlewareServer struct {
	proto.UnimplementedMiddlewareServer
	HandleFunc func(ctx context.Context, req *proto.Data) (*proto.Data, error)
}

func (m *MockMiddlewareServer) Handle(ctx context.Context, req *proto.Data) (*proto.Data, error) {
	if m.HandleFunc != nil {
		return m.HandleFunc(ctx, req)
	}
	// Default behavior: echo back the same messages with acked=true
	return &proto.Data{
		Messages: req.Messages,
	}, nil
}

// setupTestGRPCServer creates a test gRPC server with the mock middleware
func setupTestGRPCServer(mockServer *MockMiddlewareServer) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	proto.RegisterMiddlewareServer(s, mockServer)
	go func() {
		_ = s.Serve(lis) // Server will stop when the listener is closed
	}()
	return s, lis
}

// bufDialer creates a dialer function for the bufconn listener
func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestMiddlewareManager_Start_WithoutMiddleware(t *testing.T) {
	// Create channels
	inputChan := make(chan messages.Message, 10)
	outputChan := make(chan messages.Message, 10)

	// Create manager without middleware address
	manager := NewMiddlewareManager(inputChan, outputChan, "", 5, time.Second)

	// Send test messages
	testMessages := []messages.Message{
		{Id: 1, Body: []byte("test1")},
		{Id: 2, Body: []byte("test2")},
	}

	go func() {
		for _, msg := range testMessages {
			inputChan <- msg
		}
		close(inputChan)
	}()

	// Collect output messages
	var outputMessages []messages.Message
	outputDone := make(chan bool)
	go func() {
		for msg := range outputChan {
			outputMessages = append(outputMessages, msg)
		}
		outputDone <- true
	}()

	// Start manager in goroutine
	done := make(chan bool)
	go func() {
		manager.Start()
		close(outputChan)
		done <- true
	}()

	// Wait for completion
	<-done
	<-outputDone

	// Verify messages passed through unchanged
	assert.Len(t, outputMessages, 2)
	assert.Equal(t, testMessages[0].Id, outputMessages[0].Id)
	assert.Equal(t, testMessages[0].Body, outputMessages[0].Body)
	assert.Equal(t, testMessages[1].Id, outputMessages[1].Id)
	assert.Equal(t, testMessages[1].Body, outputMessages[1].Body)
}

func TestMiddlewareManager_Start_WithMiddleware_Integration(t *testing.T) {
	// Setup mock gRPC server with Unix domain socket
	mockServer := &MockMiddlewareServer{
		HandleFunc: func(ctx context.Context, req *proto.Data) (*proto.Data, error) {
			// Modify messages and set Acked=true to indicate successful processing
			responseMessages := make([]*proto.Data_Message, len(req.Messages))
			for i, msg := range req.Messages {
				responseMessages[i] = &proto.Data_Message{
					Id:    msg.Id,
					Body:  append(msg.Body, []byte("-processed")...),
					Acked: true, // Middleware indicates successful processing
				}
			}
			return &proto.Data{Messages: responseMessages}, nil
		},
	}

	// Create a temporary Unix domain socket
	socketPath := "/tmp/test_middleware_" + t.Name() + ".sock"
	// Remove socket file if it exists
	_ = os.Remove(socketPath)

	// Create Unix domain socket listener
	lis, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	defer lis.Close()
	defer os.Remove(socketPath)

	// Setup gRPC server
	s := grpc.NewServer()
	proto.RegisterMiddlewareServer(s, mockServer)

	// Start server in background
	go func() {
		_ = s.Serve(lis) // Server will stop when the listener is closed
	}()
	defer s.Stop()

	// Create channels
	inputChan := make(chan messages.Message, 10)
	outputChan := make(chan messages.Message, 10)

	// Create manager with the Unix socket address
	manager := NewMiddlewareManager(inputChan, outputChan, "unix://"+socketPath, 2, 100*time.Millisecond)

	// Send test messages
	testMessages := []messages.Message{
		{Id: 1, Body: []byte("test1")},
		{Id: 2, Body: []byte("test2")},
		{Id: 3, Body: []byte("test3")},
	}

	go func() {
		for _, msg := range testMessages {
			inputChan <- msg
		}
		close(inputChan)
	}()

	// Collect output messages
	var outputMessages []messages.Message
	outputDone := make(chan bool)
	go func() {
		for msg := range outputChan {
			outputMessages = append(outputMessages, msg)
		}
		outputDone <- true
	}()

	// Start manager in goroutine
	done := make(chan bool)
	go func() {
		manager.Start()
		close(outputChan)
		done <- true
	}()

	// Wait for completion with timeout
	select {
	case <-done:
		// Manager completed
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out")
	}

	<-outputDone

	// Verify messages were processed by middleware
	assert.Len(t, outputMessages, 3)
	assert.Equal(t, []byte("test1-processed"), outputMessages[0].Body)
	assert.Equal(t, []byte("test2-processed"), outputMessages[1].Body)
	assert.Equal(t, []byte("test3-processed"), outputMessages[2].Body)

	// Messages should NOT be acked by the middleware manager - that's the writer's job
	// But they also shouldn't be nacked since middleware processing was successful
	assert.False(t, outputMessages[0].IsAcked())
	assert.False(t, outputMessages[1].IsAcked())
	assert.False(t, outputMessages[2].IsAcked())

	// Messages should also not be nacked (since middleware indicated successful processing)
	assert.False(t, outputMessages[0].IsNacked())
	assert.False(t, outputMessages[1].IsNacked())
	assert.False(t, outputMessages[2].IsNacked())
}

func TestMiddlewareManager_GetBatch(t *testing.T) {
	t.Run("BatchSizeLimit", func(t *testing.T) {
		inputChan := make(chan messages.Message, 10)
		outputChan := make(chan messages.Message, 10)
		// Use a non-empty address to ensure this is in the "middleware available" path
		manager := NewMiddlewareManager(inputChan, outputChan, "localhost:8080", 3, 100*time.Millisecond)

		// Send 5 messages
		for i := 1; i <= 5; i++ {
			inputChan <- messages.Message{Id: uint64(i), Body: []byte("test")}
		}

		batch, ok := manager.getBatch()
		assert.True(t, ok)
		assert.Len(t, batch, 3)
		assert.Equal(t, uint64(1), batch[0].Id)
		assert.Equal(t, uint64(2), batch[1].Id)
		assert.Equal(t, uint64(3), batch[2].Id)
	})

	t.Run("Timeout", func(t *testing.T) {
		inputChan := make(chan messages.Message, 10)
		outputChan := make(chan messages.Message, 10)
		manager := NewMiddlewareManager(inputChan, outputChan, "localhost:8080", 3, 50*time.Millisecond)

		// Send only 2 messages
		inputChan <- messages.Message{Id: 1, Body: []byte("test1")}
		inputChan <- messages.Message{Id: 2, Body: []byte("test2")}

		batch, ok := manager.getBatch()
		assert.True(t, ok)
		assert.Len(t, batch, 2) // Should timeout and return partial batch
		assert.Equal(t, uint64(1), batch[0].Id)
		assert.Equal(t, uint64(2), batch[1].Id)
	})

	t.Run("ChannelCloseBeforeFirstMessage", func(t *testing.T) {
		inputChan := make(chan messages.Message, 10)
		outputChan := make(chan messages.Message, 10)
		manager := NewMiddlewareManager(inputChan, outputChan, "localhost:8080", 3, 100*time.Millisecond)

		close(inputChan)

		batch, ok := manager.getBatch()
		assert.False(t, ok)
		assert.Empty(t, batch)
	})

	t.Run("ChannelCloseDuringBatch", func(t *testing.T) {
		inputChan := make(chan messages.Message, 10)
		outputChan := make(chan messages.Message, 10)
		manager := NewMiddlewareManager(inputChan, outputChan, "localhost:8080", 3, 200*time.Millisecond)

		// Send first message and start getBatch in goroutine
		inputChan <- messages.Message{Id: 1, Body: []byte("test1")}

		batchResult := make(chan struct {
			batch []messages.Message
			ok    bool
		})

		go func() {
			batch, ok := manager.getBatch()
			batchResult <- struct {
				batch []messages.Message
				ok    bool
			}{batch, ok}
		}()

		// Give some time for getBatch to process first message and enter the timeout loop
		time.Sleep(10 * time.Millisecond)

		// Send second message and then close channel
		inputChan <- messages.Message{Id: 2, Body: []byte("test2")}
		close(inputChan)

		result := <-batchResult
		assert.False(t, result.ok)     // Channel was closed
		assert.Len(t, result.batch, 2) // Should return partial batch with messages received
		assert.Equal(t, uint64(1), result.batch[0].Id)
		assert.Equal(t, uint64(2), result.batch[1].Id)
	})
}

func TestMiddlewareManager_ProcessMessageBatch(t *testing.T) {
	// Setup mock gRPC server
	mockServer := &MockMiddlewareServer{}
	server, lis := setupTestGRPCServer(mockServer)
	defer server.Stop()

	// Create gRPC client
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewMiddlewareClient(conn)

	inputChan := make(chan messages.Message, 10)
	outputChan := make(chan messages.Message, 10)
	manager := NewMiddlewareManager(inputChan, outputChan, "", 3, time.Second)

	t.Run("SuccessfulProcessing", func(t *testing.T) {
		// Setup mock to return processed messages
		mockServer.HandleFunc = func(ctx context.Context, req *proto.Data) (*proto.Data, error) {
			responseMessages := make([]*proto.Data_Message, len(req.Messages))
			for i, msg := range req.Messages {
				responseMessages[i] = &proto.Data_Message{
					Id:    msg.Id,
					Body:  append(msg.Body, []byte("-processed")...),
					Acked: true, // Middleware indicates successful processing
				}
			}
			return &proto.Data{Messages: responseMessages}, nil
		}

		// Create test messages
		testMessages := []messages.Message{
			{Id: 1, Body: []byte("test1")},
			{Id: 2, Body: []byte("test2")},
		}

		// Process batch
		manager.processMessageBatch(testMessages, client)

		// Check output
		var outputMessages []messages.Message
		timeout := time.After(time.Second)
		for len(outputMessages) < 2 {
			select {
			case msg := <-outputChan:
				outputMessages = append(outputMessages, msg)
			case <-timeout:
				t.Fatal("Timeout waiting for output messages")
			}
		}

		assert.Len(t, outputMessages, 2)
		assert.Equal(t, []byte("test1-processed"), outputMessages[0].Body)
		assert.Equal(t, []byte("test2-processed"), outputMessages[1].Body)
		// Messages should not be acked by middleware manager
		assert.False(t, outputMessages[0].IsAcked())
		assert.False(t, outputMessages[1].IsAcked())
		// Messages should also not be nacked (successful processing)
		assert.False(t, outputMessages[0].IsNacked())
		assert.False(t, outputMessages[1].IsNacked())
	})

	t.Run("EmptyBatch", func(t *testing.T) {
		manager.processMessageBatch([]messages.Message{}, client)
		// Should not produce any output
		select {
		case <-outputChan:
			t.Fatal("Unexpected output message")
		case <-time.After(100 * time.Millisecond):
			// Expected - no output
		}
	})

	t.Run("MiddlewareNacksMessage", func(t *testing.T) {
		// Setup mock to nack specific messages
		mockServer.HandleFunc = func(ctx context.Context, req *proto.Data) (*proto.Data, error) {
			responseMessages := make([]*proto.Data_Message, len(req.Messages))
			for i, msg := range req.Messages {
				responseMessages[i] = &proto.Data_Message{
					Id:    msg.Id,
					Body:  msg.Body,
					Acked: false, // Middleware nacks this message
				}
			}
			return &proto.Data{Messages: responseMessages}, nil
		}

		// Create test message
		testMessage := messages.Message{Id: 1, Body: []byte("test1")}

		// Process batch
		manager.processMessageBatch([]messages.Message{testMessage}, client)

		// Check output
		var outputMessage messages.Message
		select {
		case outputMessage = <-outputChan:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for output message")
		}

		assert.Equal(t, []byte("test1"), outputMessage.Body)
		assert.True(t, outputMessage.IsNacked()) // Should be nacked by middleware
		assert.False(t, outputMessage.IsAcked())
	})

	t.Run("MiddlewareCallFails", func(t *testing.T) {
		// Setup mock to return an error
		mockServer.HandleFunc = func(ctx context.Context, req *proto.Data) (*proto.Data, error) {
			return nil, errors.New("middleware processing failed") // Simulate middleware failure
		}

		// Create test messages
		testMessages := []messages.Message{
			{Id: 1, Body: []byte("test1")},
			{Id: 2, Body: []byte("test2")},
		}

		// Process batch
		manager.processMessageBatch(testMessages, client)

		// Check output - all messages should be nacked
		var outputMessages []messages.Message
		timeout := time.After(time.Second)
		for len(outputMessages) < 2 {
			select {
			case msg := <-outputChan:
				outputMessages = append(outputMessages, msg)
			case <-timeout:
				t.Fatal("Timeout waiting for output messages")
			}
		}

		assert.Len(t, outputMessages, 2)
		// All messages should be nacked due to middleware failure
		assert.True(t, outputMessages[0].IsNacked())
		assert.True(t, outputMessages[1].IsNacked())
		assert.False(t, outputMessages[0].IsAcked())
		assert.False(t, outputMessages[1].IsAcked())
		// Body should remain unchanged (no processing)
		assert.Equal(t, []byte("test1"), outputMessages[0].Body)
		assert.Equal(t, []byte("test2"), outputMessages[1].Body)
	})

	t.Run("MiddlewareResponseLengthMismatch", func(t *testing.T) {
		// Setup mock to return wrong number of messages
		mockServer.HandleFunc = func(ctx context.Context, req *proto.Data) (*proto.Data, error) {
			// Return fewer messages than sent (should cause length mismatch error)
			responseMessages := []*proto.Data_Message{
				{
					Id:    req.Messages[0].Id,
					Body:  req.Messages[0].Body,
					Acked: true,
				},
				// Missing second message - this creates a length mismatch
			}
			return &proto.Data{Messages: responseMessages}, nil
		}

		// Create test messages
		testMessages := []messages.Message{
			{Id: 1, Body: []byte("test1")},
			{Id: 2, Body: []byte("test2")},
		}

		// Process batch
		manager.processMessageBatch(testMessages, client)

		// Check output - all messages should be nacked due to length mismatch
		var outputMessages []messages.Message
		timeout := time.After(time.Second)
		for len(outputMessages) < 2 {
			select {
			case msg := <-outputChan:
				outputMessages = append(outputMessages, msg)
			case <-timeout:
				t.Fatal("Timeout waiting for output messages")
			}
		}

		assert.Len(t, outputMessages, 2)
		// All messages should be nacked due to response length mismatch
		assert.True(t, outputMessages[0].IsNacked())
		assert.True(t, outputMessages[1].IsNacked())
		assert.False(t, outputMessages[0].IsAcked())
		assert.False(t, outputMessages[1].IsAcked())
		// Body should remain unchanged
		assert.Equal(t, []byte("test1"), outputMessages[0].Body)
		assert.Equal(t, []byte("test2"), outputMessages[1].Body)
	})

	t.Run("MiddlewareResponseIdMismatch", func(t *testing.T) {
		// Setup mock to return messages with wrong IDs
		mockServer.HandleFunc = func(ctx context.Context, req *proto.Data) (*proto.Data, error) {
			responseMessages := make([]*proto.Data_Message, len(req.Messages))
			for i, msg := range req.Messages {
				responseMessages[i] = &proto.Data_Message{
					Id:    msg.Id + 100, // Wrong ID - should cause mismatch error
					Body:  msg.Body,
					Acked: true,
				}
			}
			return &proto.Data{Messages: responseMessages}, nil
		}

		// Create test message
		testMessage := messages.Message{Id: 1, Body: []byte("test1")}

		// Process batch
		manager.processMessageBatch([]messages.Message{testMessage}, client)

		// Check output
		var outputMessage messages.Message
		select {
		case outputMessage = <-outputChan:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for output message")
		}

		assert.Equal(t, []byte("test1"), outputMessage.Body)
		assert.True(t, outputMessage.IsNacked()) // Should be nacked due to ID mismatch
		assert.False(t, outputMessage.IsAcked())
	})
}

func TestMiddlewareManager_IsMiddlewareNotAvailable(t *testing.T) {
	inputChan := make(chan messages.Message, 10)
	outputChan := make(chan messages.Message, 10)

	t.Run("MiddlewareNotAvailable", func(t *testing.T) {
		manager := NewMiddlewareManager(inputChan, outputChan, "", 3, time.Second)
		assert.True(t, manager.isMiddlewareNotAvailable())
	})

	t.Run("MiddlewareAvailable", func(t *testing.T) {
		manager := NewMiddlewareManager(inputChan, outputChan, "localhost:8080", 3, time.Second)
		assert.False(t, manager.isMiddlewareNotAvailable())
	})
}

func TestNewMiddlewareManager(t *testing.T) {
	inputChan := make(chan messages.Message, 10)
	outputChan := make(chan messages.Message, 10)
	address := "localhost:8080"
	batchSize := 5
	batchTimeout := 2 * time.Second

	manager := NewMiddlewareManager(inputChan, outputChan, address, batchSize, batchTimeout)

	assert.NotNil(t, manager)
	assert.Equal(t, address, manager.MiddlewareAddress)
	assert.Equal(t, batchSize, manager.BatchSize)
	assert.Equal(t, batchTimeout, manager.BatchTimeout)
	assert.Equal(t, inputChan, manager.InputChannel)
	assert.Equal(t, outputChan, manager.OutputChannel)
}
