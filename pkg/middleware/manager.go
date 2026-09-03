package middleware

import (
	"context"
	"time"

	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/softonic/homing-pigeon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog"
)

type MiddlwareManager struct {
	InputChannel      <-chan messages.Message
	OutputChannel     chan<- messages.Message
	MiddlewareAddress string
	BatchSize         int
	BatchMaxBytes     int
	BatchTimeout      time.Duration
	// pending holds a message already taken out of the input channel that was
	// left out of the batch being assembled because it would have overflowed
	// the byte budget. It opens the next batch.
	pending     *messages.Message
	CallTimeout time.Duration
}

// Start starts the middleware manager.
func (m *MiddlwareManager) Start(ctx context.Context) {
	if m.isMiddlewareNotAvailable() {
		klog.V(1).Infof("Middlewares not available")
		for {
			select {
			case <-ctx.Done():
				klog.V(1).Infof("Context cancelled, stopping middleware manager")
				return
			case msg, ok := <-m.InputChannel:
				if !ok {
					klog.V(1).Infof("Input channel closed, stopping middleware manager")
					return
				}
				m.OutputChannel <- msg
			}
		}
	}

	klog.V(1).Infof("Middlewares available")

	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(defaultMaxMessageSize), grpc.MaxCallRecvMsgSize(defaultMaxMessageSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(defaultRetryPolicy))

	conn, err := grpc.NewClient(m.MiddlewareAddress, opts...)
	if err != nil {
		klog.Errorf("fail to dial: %v", err)
	}

	klog.V(1).Infof("Middlewares connected")

	defer conn.Close()
	client := proto.NewMiddlewareClient(conn)

	for {
		select {
		case <-ctx.Done():
			// A message held over from the previous batch is no longer in the
			// input channel, so this is its last chance to be delivered.
			if m.pending != nil {
				held := *m.pending
				m.pending = nil
				m.processMessageBatch([]messages.Message{held}, client)
			}
			for {
				select {
				case msg, ok := <-m.InputChannel:
					if !ok {
						return // Channel closed
					}
					// Create single message batch and process
					m.processMessageBatch([]messages.Message{msg}, client)
				default:
					close(m.OutputChannel)
					return // No more messages, safe to exit
				}
			}
		default:
			// Do the normal work
			msgBatch, ok := m.getBatch(ctx)
			if len(msgBatch) > 0 {
				klog.V(5).Infof("Sending message to proto")
				start := time.Now()
				m.processMessageBatch(msgBatch, client)
				elapsed := time.Since(start)
				klog.V(5).Infof("Middlewares took %s", elapsed)
			}
			if !ok {
				return
			}

		}
	}

}

// tries to get a full batch of messages from the input channel or times out with the current batch size
func (m *MiddlwareManager) getBatch(ctx context.Context) ([]messages.Message, bool) {
	msgBatch := make([]messages.Message, 0, m.BatchSize)
	batchBytes := 0
	// Read the first message from the channel to avoid infinite polling timeouts when no activity
	var msg messages.Message
	var ok bool
	if m.pending != nil {
		msg, m.pending = *m.pending, nil
	} else {
		select {
		case <-ctx.Done():
			return msgBatch, false // Context cancelled
		case msg, ok = <-m.InputChannel:
			if !ok {
				return msgBatch, false // Channel closed
			}
		}
	}
	// The first message always goes in, even if it alone is over the budget:
	// it cannot be split, and holding it back would stall the pipeline.
	msgBatch = append(msgBatch, msg)
	batchBytes += len(msg.Body)
	if len(msgBatch) >= m.BatchSize {
		return msgBatch, true
	}
	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, m.BatchTimeout)
	defer cancelTimeout()
	for {
		select {
		case <-ctxTimeout.Done():
			return msgBatch, true
		case msg, ok := <-m.InputChannel:
			if !ok {
				return msgBatch, false
			}
			if m.exceedsByteBudget(batchBytes, len(msg.Body)) {
				held := msg
				m.pending = &held
				return msgBatch, true
			}
			msgBatch = append(msgBatch, msg)
			batchBytes += len(msg.Body)
			if len(msgBatch) == m.BatchSize {
				return msgBatch, true
			}
		}
	}
}

// exceedsByteBudget reports whether adding msgBytes to a batch already holding
// batchBytes would take it over BatchMaxBytes. A budget of 0 or less disables
// the check and batches are bounded by message count alone.
func (m *MiddlwareManager) exceedsByteBudget(batchBytes int, msgBytes int) bool {
	if m.BatchMaxBytes <= 0 {
		return false
	}
	return batchBytes+msgBytes > m.BatchMaxBytes
}

// send the batch of messages to the middleware and handle the responses
func (m *MiddlwareManager) processMessageBatch(msgBatch []messages.Message, client proto.MiddlewareClient) {
	if len(msgBatch) == 0 {
		return
	}
	protoMsgsRequest := make([]*proto.Data_Message, len(msgBatch))
	for i, msg := range msgBatch {
		protoMsgsRequest[i] = &proto.Data_Message{
			Id:     msg.Id,
			Body:   msg.Body,
			Acked:  msg.IsAcked(),
			Nacked: msg.IsNacked(),
		}
	}
	// send messages with wait for the middleware to be ready
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), m.CallTimeout)
	handleData, err := client.Handle(ctxTimeout, &proto.Data{
		Messages: protoMsgsRequest,
	}, grpc.WaitForReady(true))
	cancelTimeout()
	// if the middleware returns an error, we nack all messages in the batch
	if err != nil {
		for _, msg := range msgBatch {
			msg.Nack()
			m.OutputChannel <- msg
		}
		klog.Errorf("Error calling middleware %v", err)
		return
	}
	// otherwise, we handle every message in the middleware response accordingly
	protoMsgsResponse := handleData.GetMessages()
	if len(protoMsgsResponse) != len(msgBatch) {
		klog.Errorf("Middleware response length mismatch: expected %d, got %d", len(msgBatch), len(protoMsgsResponse))
		for _, msg := range msgBatch {
			msg.Nack()
			m.OutputChannel <- msg
		}
		return
	}
	for i, msg := range msgBatch {
		if msg.Id == protoMsgsResponse[i].GetId() {
			msg.Body = protoMsgsResponse[i].GetBody()
			if protoMsgsResponse[i].GetNacked() {
				msg.Nack()
			}
		} else {
			klog.Errorf("Middleware response message ID mismatch: expected %d, got %d", msg.Id, protoMsgsResponse[i].GetId())
			msg.Nack()
		}
		m.OutputChannel <- msg
	}
}

func (m *MiddlwareManager) isMiddlewareNotAvailable() bool {
	return m.MiddlewareAddress == ""
}

// NewMiddlewareManager creates a new instance of MiddlwareManager.
func NewMiddlewareManager(inputChannel chan messages.Message, outputChannel chan messages.Message, address string, batchSize int, batchTimeout time.Duration, callTimeout time.Duration, batchMaxBytes int) *MiddlwareManager {
	return &MiddlwareManager{
		InputChannel:      inputChannel,
		OutputChannel:     outputChannel,
		MiddlewareAddress: address,
		BatchSize:         batchSize,
		BatchMaxBytes:     batchMaxBytes,
		BatchTimeout:      batchTimeout,
		CallTimeout:       callTimeout,
	}
}
